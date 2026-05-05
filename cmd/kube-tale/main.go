package main

import (
	"fmt"
	"os"
)

var version = "0.0.0-dev"

func main() {
	fmt.Fprintf(os.Stdout, "kube-tale %s\n", version)
}
