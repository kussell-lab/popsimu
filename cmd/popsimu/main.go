package main

import (
	"github.com/rakyll/command"
)

func main() {
	// Register commands.
	args := []string{}

	command.On("single", "single population simulation", &cmdSinglePop{}, args)
	command.On("twopop", "two population simulation", &cmdTwoPops{}, args)

	command.ParseAndRun()
}
