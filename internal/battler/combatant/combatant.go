package combatant

import (
	"fmt"
	"slices"
	"strings"
)

type Combatant struct {
	StatBlock struct {
		Name      string         `json:"name"`
		Type      string         `json:"type"`
		HP        map[string]int `json:"hp"`
		AC        int            `json:"ac"`
		Speed     int            `json:"speed"`
		Abilities struct {
			STR int `json:"str"`
			DEX int `json:"dex"`
			CON int `json:"con"`
			INT int `json:"int"`
			WIS int `json:"wis"`
			CHA int `json:"cha"`
		} `json:"abilities"`
		Saves               map[string]int `json:"saves"`
		Skills              map[string]int `json:"skills"`
		Vulnerabilities     []string       `json:"vulnerabilities"`
		Resistances         []string       `json:"resistances"`
		Immunities          []string       `json:"immunities"`
		ConditionImmunities []string       `json:"condition_immunities"`
		Senses              map[string]int `json:"senses"`
		Languages           struct {
			Speaks      []string
			Understands []string
		} `json:"languages"`
		Traits  map[string]string `json:"traits"`
		Actions map[string]struct {
			AttackRoll struct {
				Present  bool `json:"present"`
				Modifier int  `json:"modifier"`
			} `json:"attack_roll"`
			SavingThrow struct {
				Present bool   `json:"present"`
				Ability string `json:"ability"`
				DC      int    `json:"dc"`
			} `json:"saving_throw"`
			Effects map[string]struct {
				Roll string `json:"roll"`
				Type string `json:"type"`
			} `json:"effects"`
			Description string `json:"description"`
		} `json:"actions"`
	} `json:"statblock"`
}

func (c Combatant) TakeDMG(dmg int, dmgType string) {
	if c.StatBlock.HP["current"] <= 0 {
		c.StatBlock.HP["current"] = 0
		fmt.Printf("%s was already at 0 hit points!\n", c.StatBlock.Name)
		return
	}

	if slices.Contains(c.StatBlock.Immunities, dmgType) {
		fmt.Printf("%s is immune to %s!\n", c.StatBlock.Name, dmgType)
		return
	}
	if slices.Contains(c.StatBlock.Vulnerabilities, dmgType) {
		dmg *= 2
		fmt.Printf("%s is vulnerable to %s!\n", c.StatBlock.Name, dmgType)
	}
	if slices.Contains(c.StatBlock.Resistances, dmgType) {
		dmg /= 2
		fmt.Printf("%s is resistant to %s!\n", c.StatBlock.Name, dmgType)
	}

	c.StatBlock.HP["current"] -= dmg
	if c.StatBlock.HP["current"] <= 0 {
		c.StatBlock.HP["current"] = 0
		fmt.Printf("%s has dropped to 0 hit points!\n", c.StatBlock.Name)
		return
	}
}

func (c Combatant) HealHP(hp int) {
	if c.StatBlock.HP["current"] == 0 && hp > 0 {
		fmt.Printf("%s is back above 0 hit points!\n", c.StatBlock.Name)
	}
	c.StatBlock.HP["current"] += hp
	if c.StatBlock.HP["current"] > c.StatBlock.HP["max"] {
		c.StatBlock.HP["current"] = c.StatBlock.HP["max"]
	}
}

func (c Combatant) Attack(attackRoll int) {
	if attackRoll >= c.StatBlock.AC {
		fmt.Println("Hit!")
	} else {
		fmt.Println("Miss!")
	}
}

func (c Combatant) Display() {
	sep := "-------------------------------------------------------"
	fmt.Println("=======================================================")

	fmt.Printf("%s | %s\n", capitalize(c.StatBlock.Name), c.StatBlock.Type)

	fmt.Println(sep)

	fmt.Printf(" - HP: %d/%d\n", c.StatBlock.HP["current"], c.StatBlock.HP["max"])
	fmt.Printf(" - AC: %d\n", c.StatBlock.AC)
	fmt.Printf(" - Speed: %d\n", c.StatBlock.Speed)

	fmt.Println(sep)

	fmt.Print("   STR      DEX      CON      INT      WIS      CHA   \n")
	abilityScores := []int{
		c.StatBlock.Abilities.STR,
		c.StatBlock.Abilities.DEX,
		c.StatBlock.Abilities.CON,
		c.StatBlock.Abilities.INT,
		c.StatBlock.Abilities.WIS,
		c.StatBlock.Abilities.CHA,
	}
	for _, score := range abilityScores {
		var scoreStr string
		if score >= 10 {
			scoreStr = fmt.Sprintf("  %d +%d  ", score, (score-10)/2)
		} else if score == 9 {
			scoreStr = fmt.Sprintf("  %d  +0  ", score)
		} else {
			scoreStr = fmt.Sprintf("  %d  %d  ", score, (score-10)/2)
		}
		fmt.Print(scoreStr)
	}
	fmt.Printf("\n")

	if len(c.StatBlock.Saves) != 0 {
		fmt.Println(sep)

		i := 0
		fmt.Print("Saving Throws ")
		for ability, modifier := range c.StatBlock.Saves {
			if modifier >= 0 {
				prettyPrintListItem(
					fmt.Sprintf("%s +%d", strings.ToUpper(ability), modifier),
					"",
					&i,
				)
			} else {
				prettyPrintListItem(
					fmt.Sprintf("%s %d", strings.ToUpper(ability), modifier),
					"",
					&i,
				)
			}
		}
		fmt.Printf("\n")
	}

	if len(c.StatBlock.Skills) != 0 {
		i := 0
		fmt.Print("Skills ")
		for skill, modifier := range c.StatBlock.Skills {
			if modifier >= 0 {
				prettyPrintListItem(
					fmt.Sprintf("%s +%d", capitalize(strings.Replace(skill, "_", " ", -1)), modifier),
					"",
					&i,
				)
			} else {
				prettyPrintListItem(
					fmt.Sprintf("%s %d", capitalize(strings.Replace(skill, "_", " ", -1)), modifier),
					"",
					&i,
				)
			}
		}
		fmt.Printf("\n")
	}

	printIfPopulated(c.StatBlock.Vulnerabilities, "Vulnerabilities", "")

	printIfPopulated(c.StatBlock.Resistances, "Resitances", "")

	printIfPopulated(c.StatBlock.Immunities, "Immunities", "")

	printIfPopulated(c.StatBlock.ConditionImmunities, "Condition immunities", "")

	i := 0
	fmt.Print("Senses: ")
	for sense, distance := range c.StatBlock.Senses {
		if distance > 0 {
			prettyPrintListItem(fmt.Sprintf("%s %dft", sense, distance), "", &i)
		}
	}
	passivePerception := 10
	mod, ok := c.StatBlock.Skills["perception"]
	if ok {
		passivePerception += mod
	} else {
		passivePerception += c.StatBlock.Abilities.WIS
	}
	switch i {
	case 0:
	case 1:
		fallthrough
	case 2:
		fmt.Printf(", ")
	case 3:
		fmt.Printf("\n")
	}
	fmt.Print(passivePerception)
	fmt.Printf("\n")

	fmt.Println("Languages:")

	printIfPopulated(c.StatBlock.Languages.Speaks, "-Speaks", " ")
	printIfPopulated(c.StatBlock.Languages.Understands, "-Understands", " ")

	if len(c.StatBlock.Traits) != 0 {
		fmt.Println(sep)

		fmt.Println("Traits")

		for name, trait := range c.StatBlock.Traits {
			fmt.Printf(
				"\n%s. %s\n",
				capitalize(strings.Replace(name, "_", " ", -1)),
				trait,
			)
		}
		fmt.Printf("\n")
	}

	if len(c.StatBlock.Actions) != 0 {
		fmt.Println(sep)

		fmt.Println("Actions")

		for name, action := range c.StatBlock.Actions {
			fmt.Printf(
				"\n%s. %s\n",
				capitalize(strings.Replace(name, "_", " ", -1)),
				action.Description,
			)
		}
		fmt.Printf("\n")
	}

	fmt.Println("=======================================================")

}

func prettyPrintListItem(item, indent string, i *int) {
	switch *i {
	case 0:
		fmt.Print(indent)
	case 1:
		fallthrough
	case 2:
		fmt.Printf(", ")
	case 3:
		fmt.Printf("\n")
		*i = 0
	}
	fmt.Print(item)
	*i++
}

func prettyPrintList(li []string, itemIndent string, i *int) {
	for _, item := range li {
		prettyPrintListItem(item, itemIndent, i)
	}
	fmt.Printf("\n")
}

func printIfPopulated(li []string, name, itemIndent string) {
	if len(li) != 0 {
		i := 0
		fmt.Printf("%s: ", name)
		prettyPrintList(li, itemIndent, &i)
	}
}

func capitalize(text string) string {
	words := strings.Fields(text)
	var capitalizedWords []string
	for _, word := range words {
		capitalizedWords = append(capitalizedWords, strings.ToUpper(word[:1])+word[1:])
	}
	return strings.Join(capitalizedWords, " ")
}
