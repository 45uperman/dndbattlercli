package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/45uperman/dndbattlercli/internal/battler"
	"github.com/45uperman/dndbattlercli/internal/battler/combatant"
	"github.com/45uperman/dndbattlercli/internal/battler/dice"
	"github.com/45uperman/dndbattlercli/internal/battler/spellbook"
	"github.com/45uperman/dndbattlercli/internal/process"
)

type cliCommand struct {
	name        string
	example     string
	description string
	flags       map[string]string
	callback    func(*config, []argument) error
}

type argument struct {
	text  string
	flags map[string][]string
}

type config struct {
	battler           battler.Battler
	supportedCommands map[string]cliCommand
	helpPrintList     []string
	isRunning         bool
	selection         *combatant.Combatant
}

var cfg *config

func init() {
	cfg = &config{
		supportedCommands: map[string]cliCommand{
			"exit": {
				name:        "exit",
				example:     "exit",
				description: "Exit the program",
				callback:    commandExit,
			},
			"help": {
				name:        "help",
				example:     "help",
				description: "Displays a help message",
				callback:    commandHelp,
			},
			"names": {
				name:        "names",
				example:     "names",
				description: "Displays the name of each combatant and spell stored in the battler",
				callback:    commandNames,
			},
			"select": {
				name:        "select",
				example:     "select blabby the blastoise",
				description: "Selects and displays the provided combatant",
				callback:    commandSelect,
			},
			"dmg": {
				name:        "dmg",
				example:     "dmg 28, fire",
				description: "Deals the provided amount of damage of the provided type to the selected combatant",
				callback:    commandDmg,
			},
			"heal": {
				name:        "heal",
				example:     "heal 7",
				description: "Heals the selected combatant by the provided amount of hp",
				callback:    commandHeal,
			},
			"attack": {
				name:        "attack",
				example:     "attack 18",
				description: "Compares the provided attack roll to the selected combatant's AC and displays the result",
				callback:    commandAttack,
			},
			"save": {
				name:        "save",
				example:     "save 18, dex",
				description: "Makes a saving throw using the selected comatant's saving throw modifier of the provided ability\n      against the provided DC and displays the result",
				callback:    commandSave,
			},
			"roll": {
				name:        "roll",
				example:     "roll 2d4+2",
				description: "Rolls the provided dice expression (such as d20, 8d6, or 2d4+2) and displays the total",
				flags: map[string]string{
					"--adv": "tells the battler to roll with advantage",
					"--dis": "tells the battler to roll with disadvantage",
				},
				callback: commandRoll,
			},
			"view": {
				name:        "view",
				example:     "view",
				description: "Displays the selected combatant",
				callback:    commandView,
			},
			"action": {
				name:        "action",
				example:     "action tail whip --bonus",
				description: "Takes the provided action of the selected combatant",
				flags: map[string]string{
					"--bonus": "tells the battler this is a bonus action",
					"--re":    "tells the battler this is a reaction",
				},
				callback: commandAction,
			},
			"cast": {
				name:        "cast",
				example:     "cast fireball --dc dc1 30 --am am1 19 --em em1 10, blabby the blastoise --dosav 1 1 dis --doatk 1 2 adv --do 1 3",
				description: "Casts the provided spell on the provided target(s)",
				flags: map[string]string{
					"--dc":    "tells the battler that the following DC key (dc1 in the example) should be set to the following\n   value (30 in the example)",
					"--am":    "this flag functions identically to the dc flag, but is used for attack modifiers instead\n   (+19 to hit in the example)",
					"--em":    "this flag functions identically to the dc flag, but is used for effect modifiers instead\n   (+10 to certain damage rolls in the example)",
					"--dosav": "tells the battler to force saving throw #X on the target Y times with advantage,\n   disadvantage, both (which causes them to cancel out) or neither. X is always the first field and\n   Y is always the second",
					"--doatk": "this flag functions identically to the dosav flag, but for attack rolls instead\n   (x=1 and y=2 in the example)",
					"--do":    "this flag functions identically to the dosav flag, but for unavoidable effects instead\n   (x=1 and y=3 in the example)",
				},
				callback: commandCast,
			},
		},
		helpPrintList: []string{
			"help",
			"exit",
			"roll",
			"names",
			"select",
			"view",
			"dmg",
			"heal",
			"action",
			"save",
			"cast",
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

		mappedFlags := make(map[string][]string, 1)
		for flag := range strings.SplitSeq(rawFlags, "--") {
			flagFields := strings.Fields(flag)

			if len(flagFields) == 0 {
				continue
			}

			flagName := flagFields[0]

			var flagValues []string
			if len(flagFields) < 1 {
				flagValues = []string{""}
			} else {
				flagValues = flagFields[1:]
			}

			mappedFlags[flagName] = flagValues
		}

		trimmedText := strings.TrimSpace(rawText)

		args = append(
			args,
			argument{
				text:  trimmedText,
				flags: mappedFlags,
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
	sep := "------------------------------------------------------------------------------------------------------------------------------"
	fmt.Printf("Syntax:\n%s\n", sep)
	fmt.Println("First write the name of your command, then any arguments and their respective flags and fields as follows:")
	fmt.Printf("'<command_name> <arg1> --<arg1_flag1> <arg1_flag1_field1>, <arg2> --<arg2_flag1> <arg2_flag1_field1>...'\n%s\n", sep)
	fmt.Printf("Commands:\n%s\n", sep)
	for _, commandName := range cfg.helpPrintList {
		command := cfg.supportedCommands[commandName]
		fmt.Printf("%s: %s\n\nexample: '%s'\n", command.name, command.description, command.example)
		if len(command.flags) != 0 {
			fmt.Printf("\nflags:\n")
			for flagName, flagDescription := range command.flags {
				fmt.Printf("\n%s\n\n   %s\n", flagName, flagDescription)
			}
		}
		fmt.Println(sep)
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
	c, ok := cfg.battler.GetCombatant(params[0].text)
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

	report := cfg.selection.TakeDMG(dmg, params[1].text)

	if report.WasAtZero {
		fmt.Printf("%s was already at 0 hit points!\n", cfg.selection.StatBlock.Name)
		return nil
	}
	if report.WasImmune {
		fmt.Printf("%s is immune to %s damage!\n", cfg.selection.StatBlock.Name, params[1].text)
		return nil
	}

	if report.WasResistant {
		fmt.Printf("%s is resistant to %s damage!\n", cfg.selection.StatBlock.Name, params[1].text)
	}
	if report.WasVulnerable {
		fmt.Printf("%s is vulnerable to %s damage!\n", cfg.selection.StatBlock.Name, params[1].text)
	}
	if report.DroppedToZero {
		fmt.Printf("%s dropped to 0 hit points!\n", cfg.selection.StatBlock.Name)
	}

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

	report := cfg.selection.HealHP(hp)

	if report.BackAboveZero {
		fmt.Printf("%s is back above 0 hit points!\n", cfg.selection.StatBlock.Name)
	}

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

	_, advPresent := params[1].flags["adv"]
	_, disPresent := params[1].flags["dis"]

	success, err := cfg.selection.Save(dc, ability, advPresent, disPresent)
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

	actionType := "action"
	for flagName, _ := range params[0].flags {
		switch flagName {
		case "bonus":
			actionType = "bonus action"
		case "re":
			actionType = "reaction"
		}

		if actionType != "action" {
			break
		}
	}

	actionName := strings.ReplaceAll(params[0].text, " ", "_")
	err := cfg.selection.DoAction(actionName, actionType)
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

	_, advPresent := params[0].flags["adv"]
	_, disPresent := params[0].flags["dis"]

	fmt.Println(d.Roll(advPresent, disPresent))

	return nil
}

func commandCast(cfg *config, params []argument) error {
	if len(params) < 2 {
		return fmt.Errorf("cast requires at least two arguments: the name of the spell to be cast, and the target(s)")
	}

	spellName := params[0].text
	spell, ok := cfg.battler.GetSpell(spellName)
	if !ok {
		return fmt.Errorf("spell not found: %s", spellName)
	}

	var castingLevel int
	attackModifiers := make(map[string]int, 1)
	effectModifiers := make(map[string]int, 1)
	saveDCs := make(map[string]int, 1)

	for flagName, flagValues := range params[0].flags {
		switch flagName {
		case "lvl":
			var lvl int
			_, err := fmt.Sscanf(flagValues[0], "%d", &lvl)
			if err != nil {
				continue
			}
			castingLevel = lvl
		case "am":
			if len(flagValues) < 2 {
				continue
			}
			var mod int
			_, err := fmt.Sscanf(flagValues[1], "%d", &mod)
			if err != nil {
				continue
			}
			attackModifiers[flagValues[0]] = mod
		case "em":
			if len(flagValues) < 2 {
				continue
			}
			var mod int
			_, err := fmt.Sscanf(flagValues[1], "%d", &mod)
			if err != nil {
				continue
			}
			effectModifiers[flagValues[0]] = mod
		case "dc":
			if len(flagValues) < 2 {
				continue
			}
			var dc int
			_, err := fmt.Sscanf(flagValues[1], "%d", &dc)
			if err != nil {
				continue
			}
			saveDCs[flagValues[0]] = dc
		}
	}

	spellFlags := spellbook.SpellFlags{
		CastingLevel:    castingLevel,
		AttackModifiers: attackModifiers,
		EffectModifiers: effectModifiers,
		SaveDCs:         saveDCs,
	}

	var targets []spellbook.SpellTarget

	for _, targetArgument := range params[1:] {
		c, ok := cfg.battler.GetCombatant(targetArgument.text)
		if !ok {
			continue
		}

		var doAtks []spellbook.DoEffect
		var doSavs []spellbook.DoEffect
		var doUnavoids []spellbook.DoEffect

		for flagName, flagValues := range targetArgument.flags {
			if len(flagValues) < 2 {
				continue
			}
			var id int
			_, err := fmt.Sscanf(flagValues[0], "%d", &id)
			if err != nil {
				continue
			}

			var reps int
			_, err = fmt.Sscanf(flagValues[1], "%d", &reps)
			if err != nil {
				continue
			}

			advPresent := false
			disPresent := false
			for _, value := range flagValues {
				if value == "adv" {
					advPresent = true
				}

				if value == "dis" {
					disPresent = true
				}
			}

			effectFlags := spellbook.EffectFlags{
				WithAdvantage:    advPresent,
				WithDisadvantage: disPresent,
			}

			switch flagName {
			case "doatk":
				doAtks = append(
					doAtks,
					spellbook.DoEffect{
						EffectID:    id,
						Repetitions: reps,
						Flags:       effectFlags,
					},
				)
			case "dosav":
				doSavs = append(
					doSavs,
					spellbook.DoEffect{
						EffectID:    id,
						Repetitions: reps,
						Flags:       effectFlags,
					},
				)
			case "do":
				doUnavoids = append(
					doUnavoids,
					spellbook.DoEffect{
						EffectID:    id,
						Repetitions: reps,
						Flags:       effectFlags,
					},
				)
			}
		}

		targetFlags := spellbook.TargetFlags{
			DoAttacks:      doAtks,
			DoSaves:        doSavs,
			DoUnavoidables: doUnavoids,
		}

		targets = append(
			targets,
			spellbook.SpellTarget{
				Target: c,
				Flags:  targetFlags,
			},
		)
	}

	spell.Cast(targets, spellFlags)

	return nil
}
