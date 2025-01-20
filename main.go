package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
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
	callback    func(*config) error
}

type config struct {
	next     string
	previous string
	cache    *pokecache.Cache
}

type LocationResponse struct {
	Results []struct {
		Name string `json:"name"`
	} `json:"results"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
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
	}
}

func commandExit(config *config) error {
	fmt.Printf("Closing the Pokedex... Goodbye!\n")
	os.Exit(0)
	return fmt.Errorf("error exiting")
}

func commandHelp(config *config) error {
	fmt.Printf("Welcome to the Pokedex!\n")
	fmt.Printf("Usage:\n\n")
	commands := getCommands()
	for commandName, command := range commands {
		fmt.Printf("%v: %v\n", commandName, command.description)
	}
	return nil
}

func commandMap(config *config) error {
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

func commandMapB(config *config) error {
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
			fmt.Printf("%v\n", command.callback(myConfig))
		} else {
			fmt.Printf("Unknown command\n")
		}
	}
}
