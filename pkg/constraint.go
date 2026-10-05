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
		"==": constraintEqual,
		">":  constraintGreaterThan,
		"<":  constraintLessThan,
		">=": constraintGreaterThanEqual,
		"=>": constraintGreaterThanEqual,
		"<=": constraintLessThanEqual,
		"=<": constraintLessThanEqual,
		"~":  constraintTilde,
		"^":  constraintCaret,
	}
	constraintRegexp      *regexp.Regexp
	validConstraintRegexp *regexp.Regexp
)

type operatorFunc func(v, c Version, conf conf) bool

func init() {
	ops := make([]string, 0, len(constraintOperators))
	for k := range constraintOperators {
		ops = append(ops, regexp.QuoteMeta(k))
	}

	constraintRegexp = regexp.MustCompile(fmt.Sprintf(
		`(%s)\s*(%s)`,
		strings.Join(ops, "|"),
		cvRegex))

	validConstraintRegexp = regexp.MustCompile(fmt.Sprintf(
		`^\s*(\s*(%s)\s*(%s)\s*\,?)*\s*$`,
		strings.Join(ops, "|"),
		cvRegex))
}

// Constraints is one or more constraint that a npm version can be
// checked against.
type Constraints struct {
	constraints [][]constraint
	conf        conf
}

type constraint struct {
	version  Version
	operator operatorFunc
	original string
}

// NewConstraints parses the given string and returns an instance of Constraints
func NewConstraints(v string, opts ...ConstraintOption) (Constraints, error) {
	config := new(conf)
	// Apply options
	for _, o := range opts {
		o.apply(config)
	}

	var css [][]constraint
	for _, vv := range strings.Split(v, "||") {
		// Validate the segment
		if !validConstraintRegexp.MatchString(vv) {
			return Constraints{}, xerrors.Errorf("improper constraint: %s", vv)
		}

		ss := constraintRegexp.FindAllString(vv, -1)
		if ss == nil {
			ss = append(ss, strings.TrimSpace(vv))
		}

		var cs []constraint
		for _, single := range ss {
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
		conf:        *config,
	}, nil
}

func newConstraint(c string) (constraint, error) {
	if c == "" {
		return constraint{
			version: semver.New(part.Any(true), part.Any(true), part.Any(true),
				part.NewParts("*"), ""),
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
	metadata := strings.TrimPrefix(m[9], "+")

	v := semver.New(newPart(major), newPart(minor), newPart(patch), pre, metadata)

	return constraint{
		version:  v,
		operator: constraintOperators[m[1]],
		original: c,
	}, nil
}

func newPart(p string) part.Part {
	if p == "" {
		p = "*"
	}
	return part.NewPart(p)
}

func (c constraint) check(v Version, conf conf) bool {
	op := preCheck(c.operator)
	return op(v, c.version, conf)
}

func (c constraint) String() string {
	return c.original
}

// Check tests if a version satisfies all the constraints.
func (cs Constraints) Check(v Version) bool {
	for _, c := range cs.constraints {
		if andCheck(v, c, cs.conf) {
			return true
		}
	}

	return false
}

// Returns the string format of the constraints
func (cs Constraints) String() string {
	var csStr []string
	for _, orC := range cs.constraints {
		var cstr []string
		for _, andC := range orC {
			cstr = append(cstr, andC.String())
		}
		csStr = append(csStr, strings.Join(cstr, ","))
	}

	return strings.Join(csStr, "||")
}

func andCheck(v Version, constraints []constraint, conf conf) bool {
	for _, c := range constraints {
		if !c.check(v, conf) {
			return false
		}
	}
	return true
}

// compare compares a version against a constraint version and returns -1, 0 or 1
// if the version is lower than, equal to or greater than the constraint version.
//
// Semantic Versioning ignores build metadata when determining precedence, so
// versions that differ only in metadata are equal. With the WithBuildMetadata
// option such versions are ordered by their metadata instead, following
// node-semver's compareBuild (see the option for details).
func compare(v, c Version, conf conf) int {
	result := v.Compare(c)
	if result != 0 || !conf.includeBuildMetadata || isAny(v) || isAny(c) {
		return result
	}
	return compareMetadata(v.Metadata(), c.Metadata())
}

// isAny reports whether a version carries a wildcard, i.e. it stands for a set of
// versions rather than a single one. Such a version compares equal to everything
// it covers, and build metadata must not break that tie.
// Version.IsAny only looks at the release, so the pre-release is checked as well.
func isAny(v Version) bool {
	return v.IsAny() || v.PreRelease().IsAny()
}

// sameRelease reports whether two versions share the same release, ignoring both
// the pre-release and the build metadata. It guards the pre-release branch of the
// range operators, where only "the same X.Y.Z" is asked: Release() keeps the build
// metadata, so a metadata-aware equality here would make ">1.2.3-alpha+build.1"
// reject "1.2.3-alpha+build.2".
func sameRelease(v, c Version) bool {
	return v.Release().Equal(c.Release())
}

func equal(v, c Version, conf conf) bool {
	return compare(v, c, conf) == 0
}

func greaterThan(v, c Version, conf conf) bool {
	return compare(v, c, conf) > 0
}

func greaterThanOrEqual(v, c Version, conf conf) bool {
	return compare(v, c, conf) >= 0
}

func lessThan(v, c Version, conf conf) bool {
	return compare(v, c, conf) < 0
}

func lessThanOrEqual(v, c Version, conf conf) bool {
	return compare(v, c, conf) <= 0
}

// compareMetadata orders two build metadata labels, treating a version without
// metadata as the lowest one, the way node-semver's compareBuild does.
// Semantic Versioning defines no ordering for build metadata, so node-semver
// reuses the pre-release precedence rules for it, and so does this function:
// https://github.com/npm/node-semver/blob/v7.8.5/README.md#comparison
func compareMetadata(v, c string) int {
	switch {
	case v == c:
		return 0
	case v == "":
		return -1
	case c == "":
		return 1
	}

	return newMetadataParts(v).Compare(newMetadataParts(c))
}

// newMetadataParts splits a build metadata label into comparable identifiers,
// ordered the way node-semver's compareBuild orders them: numeric identifiers
// are compared numerically and rank lower than alphanumeric ones (its
// compareIdentifiers helper), and a larger set of identifiers has a higher
// precedence, e.g. "build" < "build.1".
// part.NewPart is not used here: it maps "x" and "X" to a wildcard, while in
// build metadata they are ordinary identifiers.
//
// A numeric identifier that doesn't fit in uint64 is kept as a string, so it is
// compared lexically and ranks above every numeric one. Build numbers that large
// aren't expected in practice.
func newMetadataParts(s string) part.Parts {
	identifiers := strings.Split(s, ".")
	parts := make(part.Parts, len(identifiers))
	for i, identifier := range identifiers {
		if num, err := part.NewUint64(identifier); err == nil {
			parts[i] = num
		} else {
			parts[i] = part.NewString(identifier)
		}
	}
	return parts
}

//-------------------------------------------------------------------
// Constraint functions
//-------------------------------------------------------------------

func constraintEqual(v, c Version, conf conf) bool {
	return equal(v, c, conf)
}

func constraintGreaterThan(v, c Version, conf conf) bool {
	if !conf.includePreRelease && (c.IsPreRelease() && v.IsPreRelease()) {
		return sameRelease(v, c) && greaterThan(v, c, conf)
	}
	return greaterThan(v, c, conf)
}

func constraintLessThan(v, c Version, conf conf) bool {
	if !conf.includePreRelease && (c.IsPreRelease() && v.IsPreRelease()) {
		return sameRelease(v, c) && lessThan(v, c, conf)
	}
	return lessThan(v, c, conf)
}

func constraintGreaterThanEqual(v, c Version, conf conf) bool {
	if !conf.includePreRelease && (c.IsPreRelease() && v.IsPreRelease()) {
		return sameRelease(v, c) && greaterThanOrEqual(v, c, conf)
	}
	return greaterThanOrEqual(v, c, conf)
}

func constraintLessThanEqual(v, c Version, conf conf) bool {
	if !conf.includePreRelease && (c.IsPreRelease() && v.IsPreRelease()) {
		return sameRelease(v, c) && lessThanOrEqual(v, c, conf)
	}
	return lessThanOrEqual(v, c, conf)
}

func constraintTilde(v, c Version, conf conf) bool {
	// ~*, ~>* --> >= 0.0.0 (any)
	// ~2, ~2.x, ~2.x.x, ~>2, ~>2.x ~>2.x.x --> >=2.0.0, <3.0.0
	// ~2.0, ~2.0.x, ~>2.0, ~>2.0.x --> >=2.0.0, <2.1.0
	// ~1.2, ~1.2.x, ~>1.2, ~>1.2.x --> >=1.2.0, <1.3.0
	// ~1.2.3, ~>1.2.3 --> >=1.2.3, <1.3.0
	// ~1.2.0, ~>1.2.0 --> >=1.2.0, <1.3.0
	if !conf.includePreRelease && (c.IsPreRelease() && v.IsPreRelease()) {
		return greaterThanOrEqual(v, c, conf) && v.LessThan(c.Release())
	}
	return greaterThanOrEqual(v, c, conf) && lessThan(v, c.TildeBump(), conf)
}

func constraintCaret(v, c Version, conf conf) bool {
	// ^*      -->  (any)
	// ^1.2.3  -->  >=1.2.3 <2.0.0
	// ^1.2    -->  >=1.2.0 <2.0.0
	// ^1      -->  >=1.0.0 <2.0.0
	// ^0.2.3  -->  >=0.2.3 <0.3.0
	// ^0.2    -->  >=0.2.0 <0.3.0
	// ^0.0.3  -->  >=0.0.3 <0.0.4
	// ^0.0    -->  >=0.0.0 <0.1.0
	// ^0      -->  >=0.0.0 <1.0.0
	if !conf.includePreRelease && (c.IsPreRelease() && v.IsPreRelease()) {
		return greaterThanOrEqual(v, c, conf) && v.LessThan(c.Release())
	}
	return greaterThanOrEqual(v, c, conf) && lessThan(v, c.CaretBump(), conf)
}

func preCheck(f operatorFunc) operatorFunc {
	return func(v, c Version, conf conf) bool {
		if !conf.includePreRelease && (v.IsPreRelease() && !c.IsPreRelease()) {
			return false
		} else if c.IsPreRelease() && c.IsAny() {
			return false
		}
		return f(v, c, conf)
	}
}
