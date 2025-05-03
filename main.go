package main

import (
	"fmt"

	"github.com/45uperman/dndbattlercli/internal/process"
)

func main() {
	b, err := process.LoadFiles()
	if err != nil {
		fmt.Println(err)
		return
	}
	c := b.Combatants["Cabby the Caterpie"]
	b.DisplayNames()
	c.Display()
}
