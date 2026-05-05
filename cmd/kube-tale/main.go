// Binary kube-tale correlates Kubernetes signals into incident narratives.
package main

import (
	"fmt"
	"os"
)

var version = "0.0.0-dev"

func main() {
	if _, err := fmt.Fprintf(os.Stdout, "kube-tale %s\n", version); err != nil {
		os.Exit(1)
	}
}
