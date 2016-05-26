package main

import (
	"flag"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/mingzhi/popsimu/pop"
	"gopkg.in/yaml.v2"
)

// Config to read flags and configure file.
type cmdConfig struct {
	// Flags.
	workspace  string
	config     string
	prefix     string
	outdir     string
	ncpu       int
	numGen     int // number of generations.
	numRep     int // number of replicates.
	sampleSize int

	popConfigs []pop.Config
}

// Flags implements command package interface.
func (c *cmdConfig) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&c.workspace, "w", "", "workspace")
	fs.StringVar(&c.config, "c", "config.yaml", "configure file in YAML format.")
	fs.StringVar(&c.prefix, "p", "test", "prefix")
	fs.StringVar(&c.outdir, "o", "", "output diretory")
	fs.IntVar(&c.ncpu, "ncpu", runtime.NumCPU(), "number of CPUs for using")
	fs.IntVar(&c.numGen, "g", 1, "number of generations")
	fs.IntVar(&c.numRep, "r", 1, "number of replicates")
	fs.IntVar(&c.sampleSize, "s", 1000, "sample size")
	return fs
}

func (c *cmdConfig) ParsePopConfigs() {
	f, err := os.Open(c.config)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	content, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	if err := yaml.Unmarshal(content, &c.popConfigs); err != nil {
		panic(err)
	}
}
