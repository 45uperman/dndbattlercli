package battler

import (
	"fmt"
	"sync"

	"github.com/45uperman/dndbattlercli/internal/battler/combatant"
)

type Battler struct {
	Combatants map[string]combatant.Combatant
	MU         *sync.RWMutex
}

func (b Battler) DisplayNames() {
	count := 0
	b.MU.RLock()
	defer b.MU.RUnlock()
	for name, _ := range b.Combatants {
		switch count {
		case 0:
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

func NewBattler() Battler {
	b := Battler{
		Combatants: map[string]combatant.Combatant{},
		MU:         &sync.RWMutex{},
	}
	return b
}
