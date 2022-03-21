package main

import (
	"github.com/prairir/hotel/pkg/docker"
)

func main() {
	dock, err := docker.New()
	if err != nil {
		panic(err)
	}

	err = dock.BuildContainer("ryan", "ryan", "bruh", ".")
	if err != nil {
		panic(err)
	}

}
