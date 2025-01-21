package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/illeath/pokedex/internal/pokecache"
)

func cleanInput(text string) []string {
	lowered := strings.ToLower(text)
	trimmed := strings.TrimSpace(lowered)
	splitStr := strings.Fields(trimmed)
	return splitStr
}

type cliCommand struct {
	name        string
	description string
	callback    func(string, *config) error
}

type config struct {
	next     string
	previous string
	cache    *pokecache.Cache
	Pokedex  map[string]Pokemon
}

type LocationResponse struct {
	Results []struct {
		Name string `json:"name"`
	} `json:"results"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
}

type LocationArea struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

type Pokemon struct {
	Name            string   `json:"name"`
	Types           []string `json:"types"`
	Moves           []string `json:"moves"`
	Base_experience int      `json:"base_experience"`
	Stats           []int    `json:"stats"`
	Order           int      `json:"order"`
	Weight          int      `json:"weight"`
	Height          int      `json:"height"`
}

type PokemonEncounters struct {
	Name            string `json:"name"`
	Base_experience int    `json:"base_experience"`
	Height          int    `json:"height"`
	Weight          int    `json:"weight"`
	Types           []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
	Stats []struct {
		BaseStat int `json:"base_stat"`
		Stat     struct {
			Name string `json:"name"`
		} `json:"stat"`
	} `json:"stats"`
}

func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Explains the different commands",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "finds the next 20 locations",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "shows the previous 20 locations",
			callback:    commandMapB,
		},
		"explore": {
			name:        "explore",
			description: "explores the area",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "attempts to catch a pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "inspects a caught pokemon",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "shows the pokedex",
			callback:    commandPokedex,
		},
	}
}

func commandExit(input string, config *config) error {
	fmt.Printf("Closing the Pokedex... Goodbye!\n")
	os.Exit(0)
	return fmt.Errorf("error exiting")
}

func commandHelp(input string, config *config) error {
	fmt.Printf("Welcome to the Pokedex!\n")
	fmt.Printf("Usage:\n\n")
	commands := getCommands()
	for commandName, command := range commands {
		fmt.Printf("%v: %v\n", commandName, command.description)
	}
	return nil
}

func commandMap(input string, config *config) error {
	url := config.next
	if url == "" {
		url = "https://pokeapi.co/api/v2/location-area/"
	}
	var body []byte
	if cachedData, ok := config.cache.Get(url); ok {
		body = cachedData
	} else {
		res, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("error making GET request: %w", err)
		}
		defer res.Body.Close()

		body, err = io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %w", err)
		}
		if res.StatusCode > 299 {
			return fmt.Errorf("response failed with status code: %d and\nbody: %s", res.StatusCode, body)
		}
		config.cache.Add(url, body)
	}
	var data LocationResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Errorf("error unmarshaling json: %v", err)
	}
	for _, location := range data.Results {
		fmt.Println(location.Name)
	}
	if data.Next != nil {
		config.next = *data.Next
	} else {
		config.next = ""
	}
	if data.Previous != nil {
		config.previous = *data.Previous
	} else {
		config.previous = ""
	}
	return nil
}

func commandMapB(input string, config *config) error {
	if config.previous == "" {
		fmt.Println("you're on the first page")
		return nil
	}
	url := config.previous
	var body []byte
	if cachedData, ok := config.cache.Get(url); ok {
		body = cachedData
	} else {
		res, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("error making GET request: %w", err)
		}
		defer res.Body.Close()

		body, err = io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %w", err)
		}
		if res.StatusCode > 299 {
			return fmt.Errorf("response failed with status code: %d and\nbody: %s", res.StatusCode, body)
		}
		config.cache.Add(url, body)
	}

	var data LocationResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Errorf("error unmarshaling json: %v", err)
	}
	for _, location := range data.Results {
		fmt.Println(location.Name)
	}
	if data.Next != nil {
		config.next = *data.Next
	} else {
		config.next = ""
	}
	if data.Previous != nil {
		config.previous = *data.Previous
	} else {
		config.previous = ""
	}
	return nil
}

func commandExplore(input string, config *config) error {
	parts := strings.Fields(input)
	if len(parts) < 2 {
		return fmt.Errorf("no area specified")
	}
	areaName := parts[1]
	url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", areaName)
	var body []byte
	if cachedData, ok := config.cache.Get(url); ok {
		body = cachedData
	} else {
		res, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("error making GET request: %w", err)
		}
		defer res.Body.Close()

		body, err = io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %w", err)
		}
		if res.StatusCode > 299 {
			return fmt.Errorf("response failed with status code: %d and\nbody: %s", res.StatusCode, body)
		}
		config.cache.Add(url, body)
	}
	var data LocationArea
	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Errorf("error unmarshaling json: %v", err)
	}
	fmt.Printf("Exploring %v...\n", areaName)
	fmt.Printf("Found Pokemon:\n")
	for _, location := range data.PokemonEncounters {
		fmt.Printf(" - %v\n", location.Pokemon.Name)
	}
	return nil
}

func commandCatch(input string, config *config) error {
	if config.Pokedex == nil {
		config.Pokedex = make(map[string]Pokemon)
	}
	parts := strings.Fields(input)
	if len(parts) < 2 {
		return fmt.Errorf("error: 'catch' command must include a Pokémon name")
	}
	pokemonName := strings.ToLower(strings.TrimSpace(parts[1]))
	if pokemonName == "" {
		return fmt.Errorf("error: Pokémon name cannot be empty")
	}
	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", pokemonName)
	if input == "" || strings.TrimSpace(input) == "" {
		return fmt.Errorf("error: Pokemon name cannot be empty")
	}
	var body []byte
	if cachedData, ok := config.cache.Get(url); ok {
		body = cachedData
	} else {
		res, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("error making GET request: %w", err)
		}
		defer res.Body.Close()

		body, err = io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %w", err)
		}
		if res.StatusCode == 404 {
			return fmt.Errorf("pokemon not found: %s", input)
		} else if res.StatusCode > 299 {
			return fmt.Errorf("response failed with status code: %d and\nbody: %s", res.StatusCode, body)
		}
		config.cache.Add(url, body)
	}
	var data PokemonEncounters
	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Errorf("error unmarshaling json: %v", err)
	}
	fmt.Printf("Throwing a Pokeball at %v...\n", data.Name)
	chanceToCatch := 100 - data.Base_experience
	captured := rand.Intn(100) < chanceToCatch
	if captured {
		if _, exists := config.Pokedex[data.Name]; exists {
			fmt.Println("You already caught this Pokémon!")
		} else {
			fmt.Println("Pokemon caught")
			types := make([]string, len(data.Types))
			for i, t := range data.Types {
				types[i] = t.Type.Name
			}
			stats := make([]int, len(data.Stats))
			for i, s := range data.Stats {
				stats[i] = s.BaseStat
			}
			config.Pokedex[data.Name] = Pokemon{
				Name:            data.Name,
				Base_experience: data.Base_experience,
				Height:          data.Height,
				Weight:          data.Weight,
				Types:           types,
				Stats:           stats,
			}
		}
	} else {
		fmt.Printf("You failed to capture the pokemon")
	}
	return nil
}

func commandPokedex(input string, config *config) error {
	if len(config.Pokedex) == 0 {
		fmt.Println("Pokedex is empty")
	} else {
		fmt.Println("Pokedex:")
		for _, pokemon := range config.Pokedex {
			fmt.Printf(" - %v\n", pokemon.Name)
		}
	}
	return nil
}

func commandInspect(input string, config *config) error {
	parts := strings.Fields(input)
	if len(parts) < 2 {
		return fmt.Errorf("error: 'catch' command must include a Pokémon name")
	}
	pokemonName := strings.ToLower(strings.TrimSpace(parts[1]))
	if pokemonName == "" {
		return fmt.Errorf("error: Pokémon name cannot be empty")
	}
	if pokemon, ok := config.Pokedex[pokemonName]; !ok {
		fmt.Printf("Pokemon hasn't been caught")
	} else {
		fmt.Printf("Name: %v\n", pokemon.Name)
		fmt.Printf("Height: %v\n", pokemon.Height)
		fmt.Printf("Weight: %v\n", pokemon.Weight)
		fmt.Printf("Stats:\n")
		fmt.Printf("  -hp: %v\n", pokemon.Stats[0])
		fmt.Printf("  -attack: %v\n", pokemon.Stats[1])
		fmt.Printf("  -defense: %v\n", pokemon.Stats[2])
		fmt.Printf("  -special-attack: %v\n", pokemon.Stats[3])
		fmt.Printf("  -special-defense: %v\n", pokemon.Stats[4])
		fmt.Printf("  -speed: %v\n", pokemon.Stats[5])
		fmt.Printf("Types:\n")
		for _, t := range pokemon.Types {
			fmt.Printf("  - %v\n", t)
		}
	}
	return nil
}

func main() {
	wait := bufio.NewScanner(os.Stdin)
	commands := getCommands()
	myConfig := &config{
		next:     "",
		previous: "",
		cache:    pokecache.NewCache(5 * time.Minute),
	}
	for {
		fmt.Printf("Pokedex > ")
		wait.Scan()
		input := wait.Text()
		cleanInp := cleanInput(input)
		if command, exists := commands[cleanInp[0]]; exists {
			fmt.Printf("%v\n", command.callback(input, myConfig))
		} else {
			fmt.Printf("Unknown command\n")
		}
	}
}
