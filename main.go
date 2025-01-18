package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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
	callback    func() error
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
	}
}

func commandExit() error {
	fmt.Printf("Closing the Pokedex... Goodbye!\n")
	os.Exit(0)
	return fmt.Errorf("error exiting")
}

func commandHelp() error {
	fmt.Printf("Welcome to the Pokedex!\n")
	fmt.Printf("Usage:\n\n")
	commands := getCommands()
	for commandName, command := range commands {
		fmt.Printf("%v: %v\n", commandName, command.description)
	}
	return nil
}

func main() {
	wait := bufio.NewScanner(os.Stdin)
	commands := getCommands()
	for {
		fmt.Printf("Pokedex > ")
		wait.Scan()
		input := wait.Text()
		cleanInp := cleanInput(input)
		if command, exists := commands[cleanInp[0]]; exists {
			fmt.Printf("%v\n", command.callback())
		} else {
			fmt.Printf("Unknown command\n")
		}
	}
}
