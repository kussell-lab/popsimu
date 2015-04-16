package main

import (
	"github.com/rakyll/command"
)

func main() {
	// Register commands.
	args := []string{}

	command.On("twopop", "two population simulation", &cmdTwoPops{}, args)

	command.ParseAndRun()
}
