package dice

import (
	"fmt"
	"math/rand"
)

type Dice struct {
	Amount       int
	Denomination int
	Modifier     int
}

func (d Dice) Roll() int {
	total := 0
	for range d.Amount {
		total += (rand.Intn(d.Denomination) + 1) + d.Modifier
	}
	return total
}

var D4 = Dice{Amount: 1, Denomination: 4, Modifier: 0}
var D6 = Dice{Amount: 1, Denomination: 6, Modifier: 0}
var D8 = Dice{Amount: 1, Denomination: 8, Modifier: 0}
var D10 = Dice{Amount: 1, Denomination: 10, Modifier: 0}
var D12 = Dice{Amount: 1, Denomination: 12, Modifier: 0}
var D20 = Dice{Amount: 1, Denomination: 20, Modifier: 0}

func ReadDiceExpression(expr string) (Dice, error) {
	var amount int
	var denomination int
	var modifier int

	_, err := fmt.Sscanf(expr, "%dd%d+%d", &amount, &denomination, &modifier)
	if err != nil {
		_, err := fmt.Sscanf(expr, "%dd%d-%d", &amount, &denomination, &modifier)
		if err != nil {
			_, err := fmt.Sscanf(expr, "%dd%d", &amount, &denomination)
			if err != nil {
				_, err = fmt.Sscanf(expr, "d%d+%d", &denomination, &modifier)
				if err != nil {
					_, err = fmt.Sscanf(expr, "d%d-%d", &denomination, &modifier)
					if err != nil {
						_, err = fmt.Sscanf(expr, "d%d", &denomination)
						if err != nil {
							return Dice{}, fmt.Errorf("invalid dice expression '%s'\ntry something like '2d4+2', '8d6', or 'd20'", expr)
						}
					}
					modifier *= -1
				}
				amount = 1
			}
		}
		modifier *= -1
	}

	if amount < 0 || denomination < 0 {
		return Dice{}, fmt.Errorf("invalid dice expression '%s'\ntry something like '2d4+2', '8d6', or 'd20'", expr)
	}

	return Dice{Amount: amount, Denomination: denomination, Modifier: modifier}, nil
}
