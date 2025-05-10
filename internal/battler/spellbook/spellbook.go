package spellbook

import (
	"fmt"

	"github.com/45uperman/dndbattlercli/internal/battler/combatant"
	"github.com/45uperman/dndbattlercli/internal/battler/dice"
)

type Spell struct {
	Name               string        `json:"name"`
	Description        string        `json:"description"`
	BaseLevel          int           `json:"base_level"`
	Attacks            []SpellAttack `json:"attacks"`
	Saves              []SpellSave   `json:"saves"`
	UnavoidableEffects []SpellEffect `json:"unavoidable_effects"`
}

func (s Spell) Cast(targets []SpellTarget, spellFlags SpellFlags) {
	bigSep := "========================================================================================="
	fmt.Println(bigSep)
	fmt.Printf("%s:\n\n%s\n\n", s.Name, s.Description)
	sep := "-----------------------------------------------------------------------------------------"

	levelsAboveBase := s.BaseLevel - spellFlags.CastingLevel
	for _, target := range targets {
		fmt.Println(sep)
		fmt.Printf("TARGET: '%s'                                   NEW TARGET!\n", target.Target.StatBlock.Name)
		// Do attacks
		for _, atk := range target.Flags.DoAttacks {
			for i := range atk.Repetitions {
				fmt.Println(sep)
				fmt.Printf(
					"ATTACK: '%s', TARGET: '%s'    #%d\n\n",
					s.Attacks[atk.EffectID-1].Name,
					target.Target.StatBlock.Name,
					i+1,
				)
				s.Attacks[atk.EffectID-1].doTo(target, levelsAboveBase, spellFlags, atk.Flags, false)
			}
		}

		// Do saves
		for _, sav := range target.Flags.DoSaves {
			for i := range sav.Repetitions {
				fmt.Println(sep)
				fmt.Printf(
					"SAVE: '%s', TARGET: '%s'         #%d\n\n",
					s.Saves[sav.EffectID-1].Name,
					target.Target.StatBlock.Name,
					i+1,
				)
				err := s.Saves[sav.EffectID-1].forceOn(target, levelsAboveBase, spellFlags, sav.Flags)
				if err != nil {
					fmt.Println(err)
					continue
				}
			}
		}

		// Do unavoidable effects
		for _, unavoidable := range target.Flags.DoUnavoidables {
			for range unavoidable.Repetitions {
				result, err := s.UnavoidableEffects[unavoidable.EffectID-1].applyTo(target, levelsAboveBase, spellFlags, false)
				if err != nil {
					fmt.Println(err)
					continue
				}

				// Print
				if s.UnavoidableEffects[unavoidable.EffectID-1].EffectType == "healing" {
					fmt.Printf(
						" - Target '%s' healed %d hit points!\n",
						target.Target.StatBlock.Name,
						result,
					)
				} else {
					fmt.Printf(
						" - Target '%s' took %d %s damage!",
						target.Target.StatBlock.Name,
						result,
						s.UnavoidableEffects[unavoidable.EffectID-1].EffectType,
					)
				}
			}
		}
	}
	fmt.Println(sep)
	fmt.Println(bigSep)
}

type SpellAttack struct {
	Name             string        `json:"name"`
	ModifierKey      string        `json:"modifier_key"`
	ConditionalSaves []SpellSave   `json:"conditional_saves"`
	Effects          []SpellEffect `json:"effects"`
}

func (sa SpellAttack) doTo(target SpellTarget, levelsAboveBase int, spellFlags SpellFlags, effectFlags EffectFlags, halfEffect bool) {
	hit := target.Target.Hits(
		dice.D20.Roll(
			effectFlags.WithAdvantage,
			effectFlags.WithAdvantage,
		) + spellFlags.AttackModifiers[sa.ModifierKey],
	)

	if hit {
		fmt.Printf("Hit target '%s' with %s attack!    HIT!\n", target.Target.StatBlock.Name, sa.Name)
		// Do effects
		for _, effect := range sa.Effects {
			result, err := effect.applyTo(target, levelsAboveBase, spellFlags, halfEffect)
			if err != nil {
				fmt.Println(err)
				continue
			}

			// Print
			if effect.EffectType == "healing" {
				fmt.Printf(" - Target healed %d hit points\n", result)
			} else {
				fmt.Printf(" - Target took %d %s damage\n", result, effect.EffectType)
			}
		}
		fmt.Printf("\n")

		// Do conditional saves
		for _, save := range sa.ConditionalSaves {
			fmt.Printf(
				"CONDITIONAL SAVE: '%s', TARGET: '%s'    SAV!\n\n",
				save.Name,
				target.Target.StatBlock.Name,
			)
			err := save.forceOn(target, levelsAboveBase, spellFlags, effectFlags)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
	} else {
		fmt.Printf("Missed target '%s' with %s attack!    MISS!\n\n", target.Target.StatBlock.Name, sa.Name)
	}
}

type SpellSave struct {
	HalfEffectOnSuccess bool          `json:"half_effect_on_success"`
	Name                string        `json:"name"`
	Ability             string        `json:"ability"`
	DCKey               string        `json:"dc_key"`
	ConditionalAttacks  []SpellAttack `json:"conditional_attacks"`
	Effects             []SpellEffect `json:"effects"`
}

func (ss SpellSave) forceOn(target SpellTarget, levelsAboveBase int, spellFlags SpellFlags, effectFlags EffectFlags) error {
	saved, err := target.Target.Save(
		spellFlags.SaveDCs[ss.DCKey],
		ss.Ability,
		effectFlags.WithAdvantage,
		effectFlags.WithAdvantage,
	)
	if err != nil {
		return err
	}

	if saved {
		fmt.Printf("Target '%s' saved against %s!    SAVED!\n", target.Target.StatBlock.Name, ss.Name)
		if ss.HalfEffectOnSuccess {
			// Do effects
			for _, effect := range ss.Effects {
				result, err := effect.applyTo(target, levelsAboveBase, spellFlags, true)
				if err != nil {
					fmt.Println(err)
					continue
				}

				// Print
				if effect.EffectType == "healing" {
					fmt.Printf(" - Target still healed %d hit points\n", result)
				} else {
					fmt.Printf(" - Target still took %d %s damage\n", result, effect.EffectType)
				}
			}

			// Do conditional attacks
			for _, attack := range ss.ConditionalAttacks {
				attack.doTo(target, levelsAboveBase, spellFlags, effectFlags, true)
			}
		}
		fmt.Printf("\n")
	} else {
		fmt.Printf("Target '%s' failed it's save against %s!    FAILED!\n", target.Target.StatBlock.Name, ss.Name)

		// Do effects
		for _, effect := range ss.Effects {
			result, err := effect.applyTo(target, levelsAboveBase, spellFlags, false)
			if err != nil {
				fmt.Println(err)
				continue
			}

			// Print
			if effect.EffectType == "healing" {
				fmt.Printf(" - Target healed %d hit points\n", result)
			} else {
				fmt.Printf(" - Target took %d %s damage\n", result, effect.EffectType)
			}
		}
		fmt.Printf("\n")

		// Do conditional attacks
		for _, attack := range ss.ConditionalAttacks {
			fmt.Printf(
				"CONDITIONAL ATTACK: '%s', TARGET: '%s'    ATK\n\n",
				attack.Name,
				target.Target.StatBlock.Name,
			)
			attack.doTo(target, levelsAboveBase, spellFlags, effectFlags, false)
		}
	}

	return nil
}

type SpellEffect struct {
	ModifierKey    string `json:"modifier_key"`
	DiceExpression string `json:"dice_expression"`
	EffectType     string `json:"effect_type"`
	Upcast         Upcast `json:"upcast"`
}

func (se SpellEffect) applyTo(target SpellTarget, levelsAboveBase int, spellFlags SpellFlags, halfEffect bool) (int, error) {
	var result int

	upcastBonus, err := se.Upcast.getUpcastBonus(levelsAboveBase)
	if err != nil {
		return 0, err
	}

	result += upcastBonus

	result += spellFlags.EffectModifiers[se.ModifierKey]

	if se.DiceExpression != "" {
		effectDice, err := dice.ReadDiceExpression(se.DiceExpression)
		if err != nil {
			return 0, err
		}

		result += effectDice.Roll(false, false)
	}

	if halfEffect {
		result /= 2
	}

	var report combatant.EffectReport
	if se.EffectType == "healing" {
		report = target.Target.HealHP(result)
	} else {
		report = target.Target.TakeDMG(result, se.EffectType)
	}

	if report.WasAtZero {
		fmt.Printf(" - Target was already at 0 hit points!\n")
	}
	if report.WasImmune {
		fmt.Printf(" - Target is immune to %s damage!\n", se.EffectType)
	}
	if report.WasResistant {
		fmt.Printf(" - Target is resistant to %s damage!\n", se.EffectType)
	}
	if report.WasVulnerable {
		fmt.Printf(" - Target is vulnerable to %s damage!\n", se.EffectType)
	}
	if report.BackAboveZero {
		fmt.Printf(" - Target is back above 0 hit points!\n")
	}

	return report.TrueEffect, nil
}

type Upcast struct {
	MaxUpcast       int    `json:"max_upcast"`
	LevelsPerUpcast int    `json:"levels_per_upcast"`
	DiceExpression  string `json:"dice_expression"`
}

func (u Upcast) getUpcastBonus(levelsAboveBase int) (int, error) {
	var upcastBonus int

	if u.DiceExpression != "" {
		upcastDice, err := dice.ReadDiceExpression(u.DiceExpression)
		if err != nil {
			return 0, err
		}

		upcastLevel := min(levelsAboveBase/u.LevelsPerUpcast, u.MaxUpcast)
		for range upcastLevel {
			upcastBonus += upcastDice.Roll(false, false)
		}
	}

	return upcastBonus, nil
}

type SpellFlags struct {
	CastingLevel    int
	AttackModifiers map[string]int
	EffectModifiers map[string]int
	SaveDCs         map[string]int
}

type SpellTarget struct {
	Target *combatant.Combatant
	Flags  TargetFlags
}

type TargetFlags struct {
	DoAttacks      []DoEffect
	DoSaves        []DoEffect
	DoUnavoidables []DoEffect
}

type DoEffect struct {
	EffectID    int
	Repetitions int
	Flags       EffectFlags
}

type EffectFlags struct {
	WithAdvantage    bool
	WithDisadvantage bool
}
