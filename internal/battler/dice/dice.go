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

func ReadDiceExpression(expr string) (Dice, error) {
	var amount int
	var denomination int
	var modifier int

	_, err := fmt.Sscanf(expr, "%dd%d+%d", &amount, &denomination, &modifier)
	if err != nil {
		_, err := fmt.Sscanf(expr, "%dd%d", &amount, &denomination)
		if err != nil {
			_, err = fmt.Sscanf(expr, "d%d", &denomination)
			if err != nil {
				return Dice{}, fmt.Errorf("invalid dice expression '%s'\ntry something like '2d4+2', '8d6', or 'd20'", expr)
			}
			amount = 1
		}
	}
	return Dice{Amount: amount, Denomination: denomination, Modifier: modifier}, nil
}
