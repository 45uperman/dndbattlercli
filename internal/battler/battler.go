package battler

import (
	"fmt"
	"sync"

	"github.com/45uperman/dndbattlercli/internal/battler/combatant"
)

type Battler struct {
	Combatants map[string]combatant.Combatant
	mu         *sync.RWMutex
}

func (b Battler) DisplayNames() {
	count := 0
	b.mu.RLock()
	defer b.mu.RUnlock()
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
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Combatants[c.StatBlock.Name] = c
}

func NewBattler() Battler {
	b := Battler{
		Combatants: map[string]combatant.Combatant{},
		mu:         &sync.RWMutex{},
	}
	return b
}
