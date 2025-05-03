package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/45uperman/dndbattlercli/internal/battle"
)

func main() {
	file, err := os.Open("example_object.json")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	var c battle.Combatant
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&c)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}
	c.Display()
}
