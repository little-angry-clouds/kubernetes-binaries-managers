package main

import (
	"github.com/little-angry-clouds/kubernetes-binaries-managers/internal/wrapper"
)

func main() {
	var binName string = "kubectl"

	wrapper.Wrapper(binName)
}
