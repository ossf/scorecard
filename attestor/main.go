package main

import (
	"log"

	"github.com/ossf/scorecard-attestor/command"
)

func main() {
	if err := command.New().Execute(); err != nil {
		log.Fatalf("error during command execution: %v", err)
	}
}
