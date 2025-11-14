package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Wolfy-22/pokedex.git/internal/pokeapi"
)

type config struct {
	pokeapiClient    pokeapi.Client
	nextLocationsURL *string
	prevLocationsURL *string
}

type cliCommand struct {
	name        string
	description string
	callback    func(*config, map[string]pokeapi.Pokemon, ...string) error
}

func startRepl(cfg *config) {

	reader := bufio.NewScanner(os.Stdin)
	Pokedex, err := openPokedex()
	if err != nil {
		fmt.Println(err)
	}

	for {
		fmt.Print("Pokedex > ")
		reader.Scan()

		words := cleanInput(reader.Text())
		if len(words) == 0 {
			continue
		}

		commandName := words[0]
		args := []string{}
		if len(words) > 1 {
			args = words[1:]
		}

		command, exists := getCommands()[commandName]
		if exists {
			err := command.callback(cfg, Pokedex, args...)
			if err != nil {
				fmt.Println(err)
			}
			continue
		} else {
			fmt.Println("Unknown command")
			continue
		}
	}
}

func openPokedex() (map[string]pokeapi.Pokemon, error) {

	savedPokedex, err := os.Open(FilePath)
	if err != nil {
		return nil, fmt.Errorf("Error opening file: %v\n", err)
	}

	defer savedPokedex.Close()

	var Pokedex map[string]pokeapi.Pokemon

	decoder := json.NewDecoder(savedPokedex)
	err = decoder.Decode(&Pokedex)
	if err != nil {
		return nil, fmt.Errorf("Error decoding JSON: %v", err)
	}

	return Pokedex, nil

}

func commandExit(cfg *config, Pokedex map[string]pokeapi.Pokemon, args ...string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandSave(cfg *config, Pokedex map[string]pokeapi.Pokemon, args ...string) error {
	fmt.Println("Saving...")

	jsonData, err := json.MarshalIndent(Pokedex, "", "  ")
	if err != nil {
		return fmt.Errorf("Error marshaling JSON: %v", err)
	}

	err = os.WriteFile(FilePath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("Error writing JSON to file: %v", err)
	}

	fmt.Println("Pokedex Successfully Saved")
	return nil
}

func commandHelp(cfg *config, Pokedex map[string]pokeapi.Pokemon, args ...string) error {
	fmt.Println()
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()
	for _, cmd := range getCommands() {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}

	fmt.Println()
	return nil
}

func commandPokedex(cfg *config, Pokedex map[string]pokeapi.Pokemon, args ...string) error {
	fmt.Println()
	fmt.Println("Your Pokedex:")
	for _, pokemon := range Pokedex {
		fmt.Printf(" - %s\n", pokemon.Name)
	}
	fmt.Println()
	return nil
}

func commandMapf(cfg *config, Pokedex map[string]pokeapi.Pokemon, args ...string) error {
	fmt.Println()
	locationsResp, err := cfg.pokeapiClient.ListLocations(cfg.nextLocationsURL)
	if err != nil {
		return err
	}

	cfg.nextLocationsURL = locationsResp.Next
	cfg.prevLocationsURL = locationsResp.Previous

	for _, loc := range locationsResp.Results {
		fmt.Println(loc.Name)
	}
	fmt.Println()
	return nil
}

func commandMapb(cfg *config, Pokedex map[string]pokeapi.Pokemon, args ...string) error {
	fmt.Println()
	if cfg.prevLocationsURL == nil {
		return errors.New("you're on the first page\n")
	}

	locationResp, err := cfg.pokeapiClient.ListLocations(cfg.prevLocationsURL)
	if err != nil {
		return err
	}

	cfg.nextLocationsURL = locationResp.Next
	cfg.prevLocationsURL = locationResp.Previous

	for _, loc := range locationResp.Results {
		fmt.Println(loc.Name)
	}
	fmt.Println()
	return nil
}

func commandExplore(cfg *config, Pokedex map[string]pokeapi.Pokemon, args ...string) error {
	if len(args) != 1 {
		return errors.New("you must provide a location name")
	}

	name := args[0]
	location, err := cfg.pokeapiClient.GetLocation(name)
	if err != nil {
		return err
	}
	fmt.Printf("\nExploring %s...\n\n", location.Name)
	fmt.Println("Found Pokemon: ")
	for _, enc := range location.PokemonEncounters {
		fmt.Printf(" - %s\n", enc.Pokemon.Name)
	}
	fmt.Println()
	return nil
}

func commandCatch(cfg *config, Pokedex map[string]pokeapi.Pokemon, args ...string) error {
	if len(args) != 1 {
		return errors.New("you must provide a pokemon name")
	}

	name := args[0]
	pokemon, err := cfg.pokeapiClient.GetPokemon(name)
	if err != nil {
		return err
	}

	fmt.Printf("\nThrowing a Pokeball at %s...\n", pokemon.Name)
	fmt.Printf("%s was caught!\n\n", pokemon.Name)
	Pokedex[name] = pokemon

	return nil

}

func commandInspect(cfg *config, Pokedex map[string]pokeapi.Pokemon, args ...string) error {
	if len(args) != 1 {
		return errors.New("you must provide a pokemon name")
	}

	name := args[0]
	pokemon := Pokedex[name]

	fmt.Printf("\nName: %s\n", pokemon.Name)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)
	fmt.Printf("Base Stats:\n")
	for i := range 6 {
		fmt.Printf("  -%s: %d\n", pokemon.Stats[i].Stat.Name, pokemon.Stats[i].BaseStat)
	}
	fmt.Printf("Types:\n")
	for i := range len(pokemon.Types) {
		fmt.Printf("  - %s\n", pokemon.Types[i].Type.Name)
	}

	fmt.Println()
	return nil
}

func cleanInput(text string) []string {
	output := strings.ToLower(text)
	words := strings.Fields(output)
	return words

}

func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"explore": {
			name:        "explore <location_name>",
			description: "Explore a location",
			callback:    commandExplore,
		},
		"map": {
			name:        "map",
			description: "Get the next page of locations",
			callback:    commandMapf,
		},
		"mapb": {
			name:        "mapb",
			description: "Get the previous page of locations",
			callback:    commandMapb,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"catch": {
			name:        "catch <pokemon_name>",
			description: "Adds a Pokemon to your Pokedex",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect <pokemon_name>",
			description: "Checks a Pokemon's stats",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "Lists all caught Pokemon",
			callback:    commandPokedex,
		},
		"save": {
			name:        "save",
			description: "Saves your Pokedex",
			callback:    commandSave,
		},
	}
}
