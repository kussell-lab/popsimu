package main

import (
	"flag"
	"github.com/jacobstr/confer"
	"runtime"
	"strings"
)

// Config to read flags and configure file.
type cmdConfig struct {
	// Flags.
	workspace string // workspace, also where config file stored.
	config    string // configure file name.
	ncpu      int    // number of CPUs for using.

	// population parameters.
	popSize     int     // population size
	popNum      int     // number of populations
	genomeLen   int     // genome length.
	mutRate     float64 // mutation rate.
	inTraRate   float64 // transfer rate in the same population.
	outTraRate  float64 // transfer rate between two populations.
	fragSize    int     // size of transferred fragments.
	generations int     // number of generations to run

	// output parameters.
	outDir    string
	outPrefix string
}

// Flags implements command package interface.
func (c *cmdConfig) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&c.workspace, "w", "", "workspace")
	fs.StringVar(&c.config, "c", "config.yaml", "configure file in YAML format.")
	fs.IntVar(&c.ncpu, "ncpu", runtime.NumCPU(), "number of CPUs for using")
	return fs
}

// Parse configure file.
func (c *cmdConfig) Parse() {
	// use confer package to parse configure files.
	config := confer.NewConfig()
	// set config search root path to the workspace.
	config.SetRootPath(c.workspace)
	// read configure files.
	configPaths := strings.Split(c.config, ",")
	if err := config.ReadPaths(configPaths...); err != nil {
		panic(err)
	}
	// automatic binding.
	config.AutomaticEnv()

	// parse population parameters.
	c.popSize = config.GetInt("pop.size")
	c.popNum = config.GetInt("pop.number")
	c.generations = config.GetInt("pop.generations")
	c.genomeLen = config.GetInt("genome.length")
	c.mutRate = config.GetFloat64("mutation.rate")
	c.inTraRate = config.GetFloat64("transfer.rate.in")
	c.outTraRate = config.GetFloat64("transfer.rate.out")
	c.fragSize = config.GetInt("transfer.fragment.size")

	// parse output parameters.
	c.outDir = config.GetString("out.dir")
	c.outPrefix = config.GetString("out.prefix")
}

// Init
func (c *cmdConfig) Init() {
	runtime.GOMAXPROCS(c.ncpu)
}
