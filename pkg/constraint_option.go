package npm

type conf struct {
	includePreRelease bool
}

type ConstraintOption interface {
	apply(*conf)
}

type WithPreRelease bool

func (o WithPreRelease) apply(c *conf) {
	c.includePreRelease = bool(o)
}
