package battler

import (
	"fmt"
	"sync"

	"github.com/45uperman/dndbattlercli/internal/battler/combatant"
	"github.com/45uperman/dndbattlercli/internal/battler/spellbook"
)

type Battler struct {
	Combatants map[string]combatant.Combatant
	Spells     map[string]spellbook.Spell
	MU         *sync.RWMutex
}

func (b Battler) DisplayNames() {
	b.MU.RLock()
	defer b.MU.RUnlock()

	fmt.Println("Combatants:")
	count := 0
	for name, _ := range b.Combatants {
		switch count {
		case 0:
			fmt.Printf(" - ")
		case 1:
			fallthrough
		case 2:
			fallthrough
		case 3:
			fallthrough
		case 4:
			fmt.Printf(", ")
		case 5:
			fmt.Printf("\n")
			count = 0
		}
		fmt.Print(name)
		count++
	}

	fmt.Printf("\n")

	fmt.Println("Spells:")
	count = 0
	for name, _ := range b.Spells {
		switch count {
		case 0:
			fmt.Printf(" - ")
		case 1:
			fallthrough
		case 2:
			fallthrough
		case 3:
			fallthrough
		case 4:
			fmt.Printf(", ")
		case 5:
			fmt.Printf("\n")
			count = 0
		}
		fmt.Print(name)
		count++
	}

	fmt.Printf("\n")
}

func (b Battler) AddCombatant(c combatant.Combatant) {
	b.MU.Lock()
	defer b.MU.Unlock()
	b.Combatants[c.StatBlock.Name] = c
}

func (b Battler) AddSpell(s spellbook.Spell) {
	b.MU.Lock()
	defer b.MU.Unlock()
	b.Spells[s.Name] = s
}

func (b Battler) GetCombatant(combatantName string) (*combatant.Combatant, bool) {
	b.MU.RLock()
	defer b.MU.RUnlock()
	c, ok := b.Combatants[combatantName]
	return &c, ok
}

func (b Battler) GetSpell(spellName string) (*spellbook.Spell, bool) {
	b.MU.RLock()
	defer b.MU.RUnlock()
	s, ok := b.Spells[spellName]
	return &s, ok
}

func NewBattler() Battler {
	b := Battler{
		Combatants: map[string]combatant.Combatant{},
		Spells:     map[string]spellbook.Spell{},
		MU:         &sync.RWMutex{},
	}
	return b
}
