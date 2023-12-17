package main

import (
	"github.com/nixknight/binaries-managers/internal/wrapper"
)

func main() {
	var binName string = "kubectl"

	wrapper.Wrapper(binName)
}
