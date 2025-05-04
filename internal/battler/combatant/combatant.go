package combatant

import (
	"fmt"
	"strings"
)

type Combatant struct {
	StatBlock struct {
		Name      string         `json:"name"`
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
		Saves  []string `json:"saves"`
		Skills map[string]struct {
			Expertise bool `json:"expertise"`
		} `json:"skills"`
		Vulnerabilities     []string       `json:"vulnerabilities"`
		Resistances         []string       `json:"resistances"`
		Immunities          []string       `json:"immunities"`
		ConditionImmunities []string       `json:"condition_immunities"`
		Senses              map[string]int `json:"senses"`
		Languages           struct {
			Speaks      []string
			Understands []string
		} `json:"languages"`
	} `json:"statblock"`
}

func (c Combatant) TakeDMG(dmg int) {
	c.StatBlock.HP["current"] -= dmg
	if c.StatBlock.HP["current"] <= 0 {
		c.StatBlock.HP["current"] = 0
		fmt.Printf("%s has dropped to 0 hit points!\n", c.StatBlock.Name)
	}
}

func (c Combatant) HealHP(hp int) {
	if c.StatBlock.HP["current"] == 0 {
		fmt.Printf("%s is back above 0 hit points!\n", c.StatBlock.Name)
	}
	c.StatBlock.HP["current"] += hp
	if c.StatBlock.HP["current"] > c.StatBlock.HP["max"] {
		c.StatBlock.HP["current"] = c.StatBlock.HP["max"]
	}
}

func (c Combatant) Display() {
	sep := "-------------------------------------------------------"
	fmt.Println("=======================================================")

	trimmedText := strings.TrimSpace(c.StatBlock.Name)
	words := strings.Split(trimmedText, " ")
	for i, word := range words {
		if strings.TrimSpace(word) != "" {
			if i != 0 {
				fmt.Print(" ")
			}
			fmt.Print(strings.ToUpper(word[:1]) + word[1:])
		}
	}
	fmt.Printf("\n")

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
		} else {
			scoreStr = fmt.Sprintf("  %d  %d  ", score, (score-10)/2)
		}
		fmt.Print(scoreStr)
	}
	fmt.Printf("\n")

	fmt.Println(sep)

	printIfPopulated(c.StatBlock.Vulnerabilities, "Vulnerabilities", "")

	printIfPopulated(c.StatBlock.Resistances, "Resitances", "")

	printIfPopulated(c.StatBlock.Immunities, "Immunities", "")

	fmt.Println(sep)

	printIfPopulated(c.StatBlock.ConditionImmunities, "Condition immunities", "")

	fmt.Println(sep)

	i := 0
	fmt.Print("Senses: ")
	for sense, distance := range c.StatBlock.Senses {
		if distance > 0 {
			prettyPrintListItem(fmt.Sprintf("%s %dft", sense, distance), "", &i)
		}
	}
	fmt.Printf("\n")

	fmt.Println(sep)

	fmt.Println("Languages:")

	printIfPopulated(c.StatBlock.Languages.Speaks, "-Speaks", " ")
	printIfPopulated(c.StatBlock.Languages.Understands, "-Understands", " ")
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
