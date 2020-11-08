package main

import (
	"fmt"
	"log"

	"github.com/aquasecurity/go-npm-version/pkg"
)

func main() {
	v1, err := npm.NewVersion("1.2.3-alpha")
	if err != nil {
		log.Fatal(err)
	}

	v2, err := npm.NewVersion("1.2.3")
	if err != nil {
		log.Fatal(err)
	}

	// Comparison example. There is also GreaterThan, Equal, and just
	// a simple Compare that returns an int allowing easy >=, <=, etc.
	if v1.LessThan(v2) {
		fmt.Printf("%s is less than %s", v1, v2)
	}
}
