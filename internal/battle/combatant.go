package battle

import "fmt"

type Combatant struct {
	StatBlock struct {
		Name      string
		Stats     map[string]int  `json:"stats"`
		Abilities map[string]int  `json:"abilities"`
		Saves     map[string]bool `json:"saves"`
		Skills    map[string]struct {
			Proficiency bool `json:"proficiency"`
			Expertise   bool `json:"expertise"`
		} `json:"skills"`
		Vulnerabilities     map[string]any            `json:"vulnerabilities"`
		Resistances         map[string]any            `json:"resistances"`
		Immunities          map[string]any            `json:"immunities"`
		ConditionImmunities map[string]bool           `json:"condition_immunities"`
		Senses              map[string]int            `json:"senses"`
		Languages           map[string]map[string]any `json:"languages"`
	} `json:"combatant"`
}

func (c Combatant) Display() {
	sep := "---------------------------------------------------------------"
	fmt.Println(c.StatBlock.Name)
	fmt.Println(sep)
	for statName, statValue := range c.StatBlock.Stats {
		fmt.Printf(" -%s: %d\n", statName, statValue)
	}
	fmt.Println(sep)

	fmt.Println("Vulnerabilities:")
	err := prettyPrintAbstracts(c.StatBlock.Vulnerabilities)
	if err != nil {
		fmt.Printf("\nPrinting interrupted by error: %s", err)
		return
	}

	fmt.Println("Resistances:")
	err = prettyPrintAbstracts(c.StatBlock.Resistances)
	if err != nil {
		fmt.Printf("\nPrinting interrupted by error: %s", err)
		return
	}

	fmt.Println("Immunities:")
	err = prettyPrintAbstracts(c.StatBlock.Immunities)
	if err != nil {
		fmt.Printf("\nPrinting interrupted by error: %s", err)
		return
	}

	fmt.Println(sep)
	fmt.Println("Condition immunities:")

	i := 0
	for condition, immune := range c.StatBlock.ConditionImmunities {
		if immune {
			prettyPrintListItem(&i, condition)
		}
	}

	fmt.Println(sep)
	fmt.Println("Senses:")

	i = 0
	for sense, distance := range c.StatBlock.Senses {
		if distance > 0 {
			prettyPrintListItem(&i, sense)
		}
	}

	fmt.Println(sep)
	fmt.Println("Languages:")

	i = 0
	for k, v := range c.StatBlock.Languages {
		switch k {
		case "speaks":
			fmt.Println("-Speaks:")
			err := prettyPrintAbstracts(v)
			if err != nil {
				fmt.Printf("\nPrinting interrupted by error: %s", err)
				return
			}
		case "understands":
			fmt.Println("-Understands:")
			err := prettyPrintAbstracts(v)
			if err != nil {
				fmt.Printf("\nPrinting interrupted by error: %s", err)
				return
			}
		}
	}

}

func prettyPrintListItem(i *int, item string) {
	switch *i {
	case 0:
		fmt.Printf("  ")
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

func prettyPrintAbstracts(m map[string]any) error {
	i := 0
	for name, value := range m {
		switch v := value.(type) {
		case string:
			prettyPrintListItem(&i, name)
		case bool:
			if v {
				prettyPrintListItem(&i, name)
			}
		case []any:
			for _, name := range v {
				prettyPrintListItem(&i, name.(string))
			}
		default:
			return fmt.Errorf("prettyPrintAbstracts passed invalid object: %v", value)
		}
	}

	fmt.Printf("\n")

	return nil
}
