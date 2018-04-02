package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/mingzhi/popsimu/pop"
	"github.com/mingzhi/popsimu/simu"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	app := kingpin.New("moran-simu", "Moran population simulation")
	app.Version("0.1")

	configFile := app.Arg("config-file", "population config file").Required().String()
	outFile := app.Arg("output-file", "output file").Required().String()

	kingpin.MustParse(app.Parse(os.Args[1:]))

	pc := parsePopConfig(*configFile)
	fmt.Println(pc)
	pp := generatePopulation(pc)

	numGen := pc.Size * pc.Size * 10

	evolve(pp, pc, numGen)

	w, err := os.Create(*outFile)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(pp); err != nil {
		panic(err)
	}
}

// parsePopConfig parse a JSON PopConfig
func parsePopConfig(file string) (pc pop.Config) {
	f, err := os.Open(file)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&pc); err != nil {
		log.Fatalln(err)
	}
	return
}

func generatePopulation(pc pop.Config) *pop.Pop {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	p := pop.New()
	g := pop.NewRandomPopGenerator(r, pc.Size, pc.Length, []byte(pc.Alphabet))
	g.Operate(p)
	return p
}

func evolve(p *pop.Pop, pc pop.Config, numGen int) {
	simu.Moran([]*pop.Pop{p}, []pop.Config{pc}, numGen)
}
