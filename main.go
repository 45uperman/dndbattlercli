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
	callback    func(*config, []argument) error
}

type argument struct {
	text  string
	flags []string
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
			"save": {
				name:        "save",
				description: "Makes a saving throw using the selected comatant's saving throw modifier of the provided ability\n      against the provided DC and displays the result",
				callback:    commandSave,
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

		input := scanner.Text()
		if len(input) == 0 {
			continue
		}

		command, args, err := parseInput(input)
		if err != nil {
			fmt.Println(err)
			continue
		}

		commandStruct := cfg.supportedCommands[command]

		err = commandStruct.callback(cfg, args)
		if err != nil {
			fmt.Println(err)
		}
	}
	err = process.SaveFiles(cfg.battler)
	if err != nil {
		fmt.Println(err)
	}
}

func parseInput(input string) (command string, args []argument, err error) {
	splitInput := strings.SplitN(strings.ToLower(input), " ", 2)
	command = strings.TrimSpace(splitInput[0])
	_, ok := cfg.supportedCommands[command]
	if !ok {
		return "", []argument{}, fmt.Errorf("invalid command: '%s'", command)
	}

	if len(splitInput) == 1 {
		return command, make([]argument, 1), nil
	}

	for rawArg := range strings.SplitSeq(splitInput[1], ",") {
		rawArg := strings.TrimLeft(rawArg, " ")

		startOfFlagsIndex := strings.Index(rawArg, "--")

		if startOfFlagsIndex == -1 {
			trimmedText := strings.TrimSpace(rawArg)
			args = append(
				args,
				argument{
					text: trimmedText,
				},
			)
			continue
		}

		rawText, rawFlags := rawArg[:startOfFlagsIndex], rawArg[startOfFlagsIndex:]

		var trimmedFlags []string
		for flag := range strings.SplitSeq(rawFlags, "--") {
			trimmedFlags = append(trimmedFlags, strings.TrimSpace(flag))
		}

		trimmedText := strings.TrimSpace(rawText)

		args = append(
			args,
			argument{
				text:  trimmedText,
				flags: trimmedFlags,
			},
		)
	}

	return command, args, nil
}

func commandExit(cfg *config, params []argument) error {
	fmt.Println("Closing the program...")
	cfg.isRunning = false
	return nil
}

func commandHelp(cfg *config, params []argument) error {
	fmt.Printf("Commands:\n\n")
	for _, command := range cfg.supportedCommands {
		fmt.Printf("%s: %s\n\n", command.name, command.description)
	}
	return nil
}

func commandNames(cfg *config, params []argument) error {
	cfg.battler.DisplayNames()
	return nil
}

func commandSelect(cfg *config, params []argument) error {
	if params[0].text == "" {
		return fmt.Errorf("the select command requires the name of the combatant you want to select")
	}

	name := params[0].text
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

func commandView(cfg *config, params []argument) error {
	if cfg.selection.StatBlock.Name == "" {
		return fmt.Errorf("view requires a combatant to have already been selected using the select command")
	}

	cfg.selection.Display()
	return nil
}

func commandDmg(cfg *config, params []argument) error {
	if len(params) < 2 {
		return fmt.Errorf("dmg requires two arguments: the amount of damage, and the type of damage")
	}

	if cfg.selection.StatBlock.Name == "" {
		return fmt.Errorf("dmg requires a combatant to have already been selected using the select command")
	}

	var dmg int
	_, err := fmt.Sscanf(params[0].text, "%d", &dmg)
	if err != nil {
		return fmt.Errorf("dmg takes a whole number as it's first argument, not %s", params[0])
	}

	cfg.selection.TakeDMG(dmg, params[1].text)

	return nil
}

func commandHeal(cfg *config, params []argument) error {
	if cfg.selection.StatBlock.Name == "" {
		return fmt.Errorf("heal requires a combatant to have already been selected using the select command")
	}

	var hp int
	_, err := fmt.Sscanf(params[0].text, "%d", &hp)
	if err != nil {
		return fmt.Errorf("heal takes a whole number as an argument, not '%s'", params[0])
	}

	cfg.selection.HealHP(hp)

	return nil
}

func commandAttack(cfg *config, params []argument) error {
	if cfg.selection.StatBlock.Name == "" {
		return fmt.Errorf("attack requires a combatant to have already been selected using the select command")
	}

	var attackRoll int
	_, err := fmt.Sscanf(params[0].text, "%d", &attackRoll)
	if err != nil {
		return fmt.Errorf("attack takes a whole number as an argument, not '%s'", params[0].text)
	}

	if cfg.selection.Hits(attackRoll) {
		fmt.Println("Hit!")
	} else {
		fmt.Println("Miss!")
	}

	return nil
}

func commandSave(cfg *config, params []argument) error {
	if len(params) < 2 {
		return fmt.Errorf("save requires two arguments: the DC, and the ability to be used for the saving throw")
	}

	if cfg.selection.StatBlock.Name == "" {
		return fmt.Errorf("attack requires a combatant to have already been selected using the select command")
	}

	var dc int
	_, err := fmt.Sscanf(params[0].text, "%d", &dc)
	if err != nil {
		return fmt.Errorf("save takes a whole number as it's first argument, not '%s'", params[0].text)
	}

	ability := params[1].text

	success, err := cfg.selection.Save(dc, ability)
	if err != nil {
		return err
	}

	if success {
		fmt.Println("Success!")
	} else {
		fmt.Println("Failure!")
	}

	return nil
}

func commandAction(cfg *config, params []argument) error {
	if cfg.selection.StatBlock.Name == "" {
		return fmt.Errorf("action requires a combatant to have already been selected using the select command")
	}

	if params[0].text == "" {
		return fmt.Errorf("action takes the name of an action as it's arguments - try checking the statblock of the selected combatant\nwith the view command")
	}

	actionName := strings.ReplaceAll(params[0].text, " ", "_")
	err := cfg.selection.DoAction(actionName)
	if err != nil {
		return err
	}

	return nil
}

func commandRoll(cfg *config, params []argument) error {
	d, err := dice.ReadDiceExpression(params[0].text)
	if err != nil {
		return err
	}

	fmt.Println(d.Roll())

	return nil
}

func commandAdvantage(cfg *config, params []argument) error {
	d, err := dice.ReadDiceExpression(params[0].text)
	if err != nil {
		return err
	}

	fmt.Println(max(d.Roll(), d.Roll()))

	return nil

}

func commandDisadvantage(cfg *config, params []argument) error {
	d, err := dice.ReadDiceExpression(params[0].text)
	if err != nil {
		return err
	}

	fmt.Println(min(d.Roll(), d.Roll()))

	return nil

}
