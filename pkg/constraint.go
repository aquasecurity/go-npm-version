package npm

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/xerrors"

	"github.com/aquasecurity/go-version/pkg/part"
	"github.com/aquasecurity/go-version/pkg/semver"
)

const cvRegex string = `v?([0-9|x|X|\*]+)(\.[0-9|x|X|\*]+)?(\.[0-9|x|X|\*]+)?` +
	`(-([0-9A-Za-z\-]+(\.[0-9A-Za-z\-]+)*))?` +
	`(\+([0-9A-Za-z\-]+(\.[0-9A-Za-z\-]+)*))?`

var (
	constraintOperators = map[string]operatorFunc{
		"":   constraintEqual,
		"=":  constraintEqual,
		">":  constraintGreaterThan,
		"<":  constraintLessThan,
		">=": constraintGreaterThanEqual,
		"=>": constraintGreaterThanEqual,
		"<=": constraintLessThanEqual,
		"=<": constraintLessThanEqual,
		"~":  constraintTilde,
		"^":  constraintCaret,
	}
	constraintRegexp *regexp.Regexp
)

type operatorFunc func(v, c Version) bool

func init() {
	ops := make([]string, 0, len(constraintOperators))
	for k := range constraintOperators {
		ops = append(ops, regexp.QuoteMeta(k))
	}

	constraintRegexp = regexp.MustCompile(fmt.Sprintf(
		`^\s*(%s)\s*(%s)\s*$`,
		strings.Join(ops, "|"),
		cvRegex))
}

type Constraints struct {
	constraints [][]constraint
}

// Constraints is one or more constraint that a npm version can be
// checked against.
type constraint struct {
	version  Version
	operator operatorFunc
}

// NewConstraints parses the given string and returns an instance of Constraints
func NewConstraints(v string) (Constraints, error) {
	var css [][]constraint
	for _, vv := range strings.Split(v, "||") {
		var cs []constraint
		for _, single := range strings.Split(vv, ",") {
			c, err := newConstraint(single)
			if err != nil {
				return Constraints{}, err
			}
			cs = append(cs, c)
		}
		css = append(css, cs)
	}

	return Constraints{
		constraints: css,
	}, nil

}

func newConstraint(c string) (constraint, error) {
	if c == "" {
		return constraint{
			version: Version{
				Version: semver.New(part.Any(true), part.Any(true), part.Any(true),
					part.NewParts("*"), ""),
			},
			operator: constraintOperators[""],
		}, nil
	}

	m := constraintRegexp.FindStringSubmatch(c)
	if m == nil {
		return constraint{}, xerrors.Errorf("improper constraint: %s", c)
	}

	major := m[3]
	minor := strings.TrimPrefix(m[4], ".")
	patch := strings.TrimPrefix(m[5], ".")
	pre := part.NewParts(strings.TrimPrefix(m[6], "-"))

	v := semver.New(newPart(major), newPart(minor), newPart(patch), pre, "")

	return constraint{
		version:  Version{v},
		operator: constraintOperators[m[1]],
	}, nil
}

func newPart(p string) part.Part {
	if p == "" {
		p = "*"
	}
	return part.NewPart(p)
}

func (c constraint) check(v Version) bool {
	op := preCheck(c.operator)
	return op(v, c.version)
}

// Check tests if a version satisfies all the constraints.
func (cs Constraints) Check(v Version) bool {
	for _, c := range cs.constraints {
		if andCheck(v, c) {
			return true
		}
	}

	return false
}

func andCheck(v Version, constraints []constraint) bool {
	for _, c := range constraints {
		if !c.check(v) {
			return false
		}
	}
	return true
}

//-------------------------------------------------------------------
// Constraint functions
//-------------------------------------------------------------------

func constraintEqual(v, c Version) bool {
	return v.Equal(c.Version)
}

func constraintGreaterThan(v, c Version) bool {
	if c.IsPreRelease() && v.IsPreRelease() {
		return v.Release().Equal(c.Release()) && v.GreaterThan(c.Version)
	}
	return v.GreaterThan(c.Version)
}

func constraintLessThan(v, c Version) bool {
	if c.IsPreRelease() && v.IsPreRelease() {
		return v.Release().Equal(c.Release()) && v.LessThan(c.Version)
	}
	return v.LessThan(c.Version)
}

func constraintGreaterThanEqual(v, c Version) bool {
	if c.IsPreRelease() && v.IsPreRelease() {
		return v.Release().Equal(c.Release()) && v.GreaterThanOrEqual(c.Version)
	}
	return v.GreaterThanOrEqual(c.Version)
}

func constraintLessThanEqual(v, c Version) bool {
	if c.IsPreRelease() && v.IsPreRelease() {
		return v.Release().Equal(c.Release()) && v.LessThanOrEqual(c.Version)
	}
	return v.LessThanOrEqual(c.Version)
}

func constraintTilde(v, c Version) bool {
	// ~*, ~>* --> >= 0.0.0 (any)
	// ~2, ~2.x, ~2.x.x, ~>2, ~>2.x ~>2.x.x --> >=2.0.0, <3.0.0
	// ~2.0, ~2.0.x, ~>2.0, ~>2.0.x --> >=2.0.0, <2.1.0
	// ~1.2, ~1.2.x, ~>1.2, ~>1.2.x --> >=1.2.0, <1.3.0
	// ~1.2.3, ~>1.2.3 --> >=1.2.3, <1.3.0
	// ~1.2.0, ~>1.2.0 --> >=1.2.0, <1.3.0
	if c.IsPreRelease() && v.IsPreRelease() {
		return v.GreaterThanOrEqual(c.Version) && v.LessThan(c.Release())
	}
	return v.GreaterThanOrEqual(c.Version) && v.LessThan(c.TildeBump())
}

func constraintCaret(v, c Version) bool {
	// ^*      -->  (any)
	// ^1.2.3  -->  >=1.2.3 <2.0.0
	// ^1.2    -->  >=1.2.0 <2.0.0
	// ^1      -->  >=1.0.0 <2.0.0
	// ^0.2.3  -->  >=0.2.3 <0.3.0
	// ^0.2    -->  >=0.2.0 <0.3.0
	// ^0.0.3  -->  >=0.0.3 <0.0.4
	// ^0.0    -->  >=0.0.0 <0.1.0
	// ^0      -->  >=0.0.0 <1.0.0
	if c.IsPreRelease() && v.IsPreRelease() {
		return v.GreaterThanOrEqual(c.Version) && v.LessThan(c.Release())
	}
	return v.GreaterThanOrEqual(c.Version) && v.LessThan(c.CaretBump())
}

func preCheck(f operatorFunc) operatorFunc {
	return func(v, c Version) bool {
		if v.IsPreRelease() && !c.IsPreRelease() {
			return false
		} else if c.IsPreRelease() && c.IsAny() {
			return false
		}
		return f(v, c)
	}
}
