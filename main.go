package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/45uperman/dndbattlercli/internal/battler"
	"github.com/45uperman/dndbattlercli/internal/battler/combatant"
	"github.com/45uperman/dndbattlercli/internal/battler/dice"
	"github.com/45uperman/dndbattlercli/internal/process"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*config, []string) error
}

type config struct {
	battler           battler.Battler
	supportedCommands map[string]cliCommand
	isRunning         bool
	selection         combatant.Combatant
}

var cfg *config

func init() {
	cfg = &config{
		supportedCommands: map[string]cliCommand{
			"exit": {
				name:        "exit",
				description: "Exit the program",
				callback:    commandExit,
			},
			"help": {
				name:        "help",
				description: "Displays a help message",
				callback:    commandHelp,
			},
			"names": {
				name:        "names",
				description: "Displays the name of each combatant stored in the battler",
				callback:    commandNames,
			},
			"select": {
				name:        "select",
				description: "Selects and displays the provided combatant",
				callback:    commandSelect,
			},
			"dmg": {
				name:        "dmg",
				description: "Deals the provided amount of damage to the selected combatant",
				callback:    commandDmg,
			},
			"heal": {
				name:        "heal",
				description: "Heals the selected combatant by the provided amount of hp",
				callback:    commandHeal,
			},
			"attack": {
				name:        "attack",
				description: "Compares the provided attack roll to the selected combatant's AC and displays the result",
				callback:    commandAttack,
			},
			"roll": {
				name:        "roll",
				description: "Rolls the provided amount of dice of the provided denomination and displays the total",
				callback:    commandRoll,
			},
			"advantage": {
				name:        "advantage",
				description: "Rolls the provided amount of dice of the provided denomination twice and displays the highest total",
				callback:    commandAdvantage,
			},
			"disadvantage": {
				name:        "disadvantage",
				description: "Rolls the provided amount of dice of the provided denomination twice and displays the lowest total",
				callback:    commandDisadvantage,
			},
			"view": {
				name:        "view",
				description: "Displays the selected combatant",
				callback:    commandView,
			},
			"action": {
				name:        "action",
				description: "Takes the provided action of the selected combatant",
				callback:    commandAction,
			},
		},
		isRunning: true,
	}

	b, err := process.LoadFiles()
	if err != nil {
		fmt.Println(err)
		return
	}
	cfg.battler = b
}

func main() {
	var err error
	scanner := bufio.NewScanner(os.Stdin)
	for cfg.isRunning {
		fmt.Print("D&DBattler > ")

		ok := scanner.Scan()
		if !ok {
			fmt.Println(scanner.Err())
			cfg.isRunning = false
			continue
		}

		input := cleanInput(scanner.Text())
		if len(input) == 0 {
			continue
		}

		command := input[0]
		var args []string
		if len(input) == 1 {
			args = []string{""}
		} else {
			args = input[1:]
		}

		commandStruct, ok := cfg.supportedCommands[command]
		if !ok {
			fmt.Println("Invalid command!")
			continue
		}

		err = commandStruct.callback(cfg, args)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func cleanInput(text string) (cleanWords []string) {
	return strings.Fields(strings.ToLower(text))
}

func commandExit(cfg *config, params []string) error {
	fmt.Println("Closing the program...")
	cfg.isRunning = false
	return nil
}

func commandHelp(cfg *config, params []string) error {
	fmt.Printf("Commands:\n\n")
	for _, command := range cfg.supportedCommands {
		fmt.Printf("%s: %s\n", command.name, command.description)
	}
	return nil
}

func commandNames(cfg *config, params []string) error {
	cfg.battler.DisplayNames()
	return nil
}

func commandSelect(cfg *config, params []string) error {
	if params[0] == "" {
		return fmt.Errorf("the select command requires the name of the combatant you want to select")
	}

	name := strings.Join(params, " ")
	cfg.battler.MU.RLock()
	defer cfg.battler.MU.RUnlock()
	c, ok := cfg.battler.Combatants[name]
	if !ok {
		return fmt.Errorf("could not find combatant: %s", name)
	}
	cfg.selection = c
	fmt.Println("Selection:")
	c.Display()
	return nil
}

func commandView(cfg *config, params []string) error {
	if cfg.selection.StatBlock.Name == "" {
		return fmt.Errorf("view requires a combatant to have already been selected using the select command")
	}

	cfg.selection.Display()
	return nil
}

func commandDmg(cfg *config, params []string) error {
	if len(params) < 2 {
		return fmt.Errorf("dmg requires two arguments: the amount of damage, and the type of damage")
	}

	if cfg.selection.StatBlock.Name == "" {
		return fmt.Errorf("dmg requires a combatant to have already been selected using the select command")
	}

	var dmg int
	_, err := fmt.Sscanf(params[0], "%d", &dmg)
	if err != nil {
		return fmt.Errorf("dmg takes a whole number as it's first argument, not %s", params[0])
	}

	cfg.selection.TakeDMG(dmg, params[1])

	return nil
}

func commandHeal(cfg *config, params []string) error {
	if cfg.selection.StatBlock.Name == "" {
		return fmt.Errorf("heal requires a combatant to have already been selected using the select command")
	}

	var hp int
	_, err := fmt.Sscanf(params[0], "%d", &hp)
	if err != nil {
		return fmt.Errorf("heal takes a whole number as an argument, not '%s'", params[0])
	}

	cfg.selection.HealHP(hp)

	return nil
}

func commandAttack(cfg *config, params []string) error {
	if cfg.selection.StatBlock.Name == "" {
		return fmt.Errorf("attack requires a combatant to have already been selected using the select command")
	}

	var attackRoll int
	_, err := fmt.Sscanf(params[0], "%d", &attackRoll)
	if err != nil {
		return fmt.Errorf("attack takes a whole number as an argument, not '%s'", params[0])
	}

	cfg.selection.Attack(attackRoll)

	return nil
}

func commandAction(cfg *config, params []string) error {
	if cfg.selection.StatBlock.Name == "" {
		return fmt.Errorf("action requires a combatant to have already been selected using the select command")
	}

	if params[0] == "" {
		return fmt.Errorf("action takes the name of an action as it's arguments - try checking the statblock of the selected combatant\nwith the view command")
	}

	err := cfg.selection.DoAction(strings.Join(params, "_"))
	if err != nil {
		return err
	}

	return nil
}

func commandRoll(cfg *config, params []string) error {
	d, err := dice.ReadDiceExpression(params[0])
	if err != nil {
		return err
	}

	fmt.Println(d.Roll())

	return nil
}

func commandAdvantage(cfg *config, params []string) error {
	d, err := dice.ReadDiceExpression(params[0])
	if err != nil {
		return err
	}

	fmt.Println(max(d.Roll(), d.Roll()))

	return nil

}

func commandDisadvantage(cfg *config, params []string) error {
	d, err := dice.ReadDiceExpression(params[0])
	if err != nil {
		return err
	}

	fmt.Println(min(d.Roll(), d.Roll()))

	return nil

}
