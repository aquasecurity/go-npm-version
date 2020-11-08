package main

import (
	"fmt"
	"log"

	"github.com/aquasecurity/go-npm-version/pkg"
)

func main() {
	v, err := npm.NewVersion("2.1.0")
	if err != nil {
		log.Fatal(err)
	}

	c, err := npm.NewConstraints(">= 1.0, < 1.4 || > 2.0")
	if err != nil {
		log.Fatal(err)
	}

	if c.Check(v) {
		fmt.Printf("%s satisfies constraints '%s'", v, c)
	}
}
