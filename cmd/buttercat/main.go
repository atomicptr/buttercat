package main

import (
	"log"

	"github.com/atomicptr/buttercat/pkg/cli"
)

func main() {
	err := cli.Run()
	if err != nil {
		log.Fatal(err)
	}
}
