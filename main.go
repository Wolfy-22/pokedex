package main

import (
	"time"

	"github.com/Wolfy-22/pokedex.git/internal/pokeapi"
)

const (
	FilePath = "./internal/Saves/save.json"
)

func main() {
	pokeClient := pokeapi.NewClient(5*time.Second, time.Minute*5)
	cfg := &config{
		pokeapiClient: pokeClient,
	}

	startRepl(cfg)
}
