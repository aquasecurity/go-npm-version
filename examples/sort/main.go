package main

import (
	"fmt"
	"log"
	"sort"

	npm "github.com/aquasecurity/go-npm-version/pkg"
)

func main() {
	versionsRaw := []string{"1.1.0", "0.7.1", "1.4.0-alpha", "1.4.0-beta", "1.4.0", "1.4.0-alpha.1"}
	versions := make([]npm.Version, len(versionsRaw))
	for i, raw := range versionsRaw {
		v, err := npm.NewVersion(raw)
		if err != nil {
			log.Fatal(err)
		}
		versions[i] = v
	}

	// After this, the versions are properly sorted
	sort.Sort(npm.Collection(versions))

	for _, v := range versions {
		fmt.Println(v)
	}
}
