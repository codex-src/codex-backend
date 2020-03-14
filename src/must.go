package main

import (
	"fmt"
	"log"
)

func must(err error, desc string) {
	if err == nil {
		// No-op
		return
	}
	log.Fatal(fmt.Errorf("%s: %w", desc, err))
}
