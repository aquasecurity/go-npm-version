# go-npm-version

![Test](https://github.com/aquasecurity/go-npm-version/workflows/Test/badge.svg?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/aquasecurity/go-npm-version)](https://goreportcard.com/report/github.com/aquasecurity/go-npm-version)
![GitHub](https://img.shields.io/github/license/aquasecurity/go-npm-version)

go-npm-version is a library for parsing npm versions and version constraints, and verifying versions against a set of constraints.
go-npm-version can sort a collection of versions properly, handles prerelease versions, etc.

Versions used with go-npm-version must follow [Semantic Versioning](https://semver.org/).
Constraints used with go-npm-version must follow [npm rules](https://nodejs.dev/learn/semantic-versioning-using-npm).

For more details, see [here](https://docs.npmjs.com/cli/v6/using-npm/semver)

## Usage
### Version Parsing and Comparison

See [example](./examples/comparison/main.go)

```
v1, _ := npm.NewVersion("1.2.3-alpha")
v2, _ := npm.NewVersion("1.2.3")

// Comparison example. There is also GreaterThan, Equal, and just
// a simple Compare that returns an int allowing easy >=, <=, etc.
if v1.LessThan(v2) {
	fmt.Printf("%s is less than %s", v1, v2)
}
```

### Version Constraints
See [example](./examples/constraint/main.go)

```
v, _ := npm.NewVersion("2.1.0")
c, _ := npm.NewConstraints(">= 1.0, < 1.4 || > 2.0")

if c.Check(v) {
	fmt.Printf("%s satisfies constraints '%s'", v, c)
}
```

### Version Sorting
See [example](./examples/sort/main.go)

```
versionsRaw := []string{"1.1.0", "0.7.1", "1.4.0-alpha", "1.4.0-beta", "1.4.0", "1.4.0-alpha.1"}
versions := make([]npm.Version, len(versionsRaw))
for i, raw := range versionsRaw {
	v, _ := npm.NewVersion(raw)
	versions[i] = v
}

// After this, the versions are properly sorted
sort.Sort(npm.Collection(versions))
```

### CLI

#### Build
```
$ go build -o version cmd/version/main.go
```

#### Compare

```
$ ./version compare 0.1.2 0.1.3
-1
```

#### Constraint

```
$ ./version satisfy 0.1.2 ">0.1.1"
true
```

## Status
go-npm-version doesn't support a range of versions yet.

- [x] `^` (e.g. ^0.13.0)
- [x] `~` (e.g. ~0.13.0)
- [x] `>`
- [x] `>=`
- [x] `<`
- [x] `<=`
- [x] `=`
- [ ] `-` (e.g. 2.1.0 - 2.6.2)
- [x] `||` (e.g. < 2.1 || > 2.6)
