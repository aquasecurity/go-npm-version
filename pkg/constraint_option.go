package npm

type conf struct {
	includePreRelease    bool
	includeBuildMetadata bool
}

type ConstraintOption interface {
	apply(*conf)
}

type WithPreRelease bool

func (o WithPreRelease) apply(c *conf) {
	c.includePreRelease = bool(o)
}

// WithBuildMetadata makes build metadata significant when checking constraints.
// Semantic Versioning requires it to be ignored and defines no ordering for it,
// so by default "1.2.3+build.1" and "1.2.3+build.2" are equal. node-semver does
// define an ordering in compareBuild, which backs its sort and rsort, and this
// option applies that ordering to constraint checks:
// https://github.com/npm/node-semver/blob/v7.8.5/README.md#comparison
//
//	"1.2.3" < "1.2.3+build" < "1.2.3+build.1" < "1.2.3+build.9" <
//	"1.2.3+build.10" < "1.2.3+build.alpha"
//
// It applies to constraints that carry no metadata as well, so "1.2.3+build.1"
// no longer satisfies "=1.2.3" and does satisfy ">1.2.3". Constraints with a
// wildcard (e.g. "2", "1.2.x" or "1.2.3-x") still ignore build metadata.
// Version comparison (LessThan, Equal, Compare) is not affected and always
// follows Semantic Versioning.
type WithBuildMetadata bool

func (o WithBuildMetadata) apply(c *conf) {
	c.includeBuildMetadata = bool(o)
}
