# D&D Battler
## What It's For
This is a simple CLI made to automate a lot of the more tedious, math-y elements of combat in 5e,
especially for DMs running encounters with lots of enemies. There are already lots of simpler
ways to do this, but I've been wanting to make (and use) something like this for quite some time.
## Why It Exists
Like I said, I've been wanting to make something like this for a pretty long time now, but the
reason I finally got around to it was boot.dev. This is a personal project for boot.dev, and
for that reason, this has actually been a bit rushed. I haven't gotten around to writing any
test for it (spare me), and I'm sure there are *lots* of bugs I haven't found yet just testing
manually.
## How To Use It
There are two main kinds of objects that this program interacts with; Combatants, and Spells.
I'll go into more detail about both soon, but it shouldn't be *too* hard to figure out if
you're familiar at all with JSON. I've included a couple of example combatants in the
battle_files/combatants directory, which are just .json files with the objects inside. If you
want to make your own comatants, just make a new .json file and either write the whole thing out
yourself, or copy the objects and replace or omit fields as needed.
### Combatants
In case the structure isn't
clear, I'll show the Go structs from the internal/battler/combatant/combatant.go file here:
```
type Combatant struct {
	StatBlock struct {
		FileName  string         `json:"file_name"`
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
		Traits       map[string]string `json:"traits"`
		Actions      map[string]Action `json:"actions"`
		BonusActions map[string]Action `json:"bonus_actions"`
		Reactions    map[string]Action `json:"reactions"`
	} `json:"statblock"`
}
```
As well as the relevant Action struct:
```
type Action struct {
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
}
```
### Spells
All that about copying .json files and replacing fields goes for spells, too. You can
find an example spell in battle_files/spells, and that's where you'll need to put any
new ones you want to create. Here are the relevant Go structs from the
internal/battler/spellbook/spellbook.go file:
```
type Spell struct {
	Name               string        `json:"name"`
	Description        string        `json:"description"`
	BaseLevel          int           `json:"base_level"`
	Attacks            []SpellAttack `json:"attacks"`
	Saves              []SpellSave   `json:"saves"`
	UnavoidableEffects []SpellEffect `json:"unavoidable_effects"`
}

type SpellAttack struct {
	Name             string        `json:"name"`
	ModifierKey      string        `json:"modifier_key"`
	ConditionalSaves []SpellSave   `json:"conditional_saves"`
	Effects          []SpellEffect `json:"effects"`
}

type SpellSave struct {
	HalfEffectOnSuccess bool          `json:"half_effect_on_success"`
	Name                string        `json:"name"`
	Ability             string        `json:"ability"`
	DCKey               string        `json:"dc_key"`
	ConditionalAttacks  []SpellAttack `json:"conditional_attacks"`
	Effects             []SpellEffect `json:"effects"`
}

type SpellEffect struct {
	ModifierKey    string `json:"modifier_key"`
	DiceExpression string `json:"dice_expression"`
	EffectType     string `json:"effect_type"`
	Upcast         Upcast `json:"upcast"`
}
```
As long as you get the names, values, and JSON syntax right, everything *should* work fine.
### Commands
The **help** command should be good enough explanation, but even I forget how exactly the
**cast** command works each time I go to use it, so I'll try my best to further clarify
that one here. First off, commands in this program have this structure:
```
<command_name> <arg1> --<arg1_flag1> <arg1_flag1_field1>, <arg2> --<arg2_flag1> <arg2_flag1_field1>...
```
A command name, like **cast** followed by whitespace, then arugments separated by commas with
flags marked by two dashes `--`. Each flag can have any number of fields, each one separated by
whitespace. *Most* commands do not user flags, or even multiple arguments. Only **dmg**, **save**,
**action**, and **cast** care about more than one argument. Only **roll**, **action**, and **cast**
care about flags. Any unnecessary arguments, flags, and flag fields are ignored. Now, about the **cast**
command: The **cast** command is the most complicated by far, and is the main reason I added flags to
any of the commands. Here is the example provided in the **help** command:
```
cast fireball --dc dc1 30 --am am1 19 --em em1 10, blabby the blastoise --dosav 1 1 dis --doatk 1 2 adv --do 1 3
```
There's a lot going on here, but it's not *that* bad. The first field of --dc, --am, and --em is
a key, which different spell effects will use differently. A SpellSave, for example, needs a DC. It
gets this DC by looking at it's own predefined DCKey field and searching for it in the flags you write out.
In the examples, the DCKey for all the SpellSaves is `dc1`, so the example command sets it to `30`. `am1`
is then set to `19`, and `em1` is set to `10`. So the spell DC will be `30`, the spell attacks involved
will have a +`19` to hit, and the spell deal an additional `10` damage (or healing) with each of it's
effects. A *cure wounds* spell, for example, needs an effect modifier, typically your Cleric or Druid's
Wisdom modifier, to add to it's 1d8 healing. `--em em1 {your_cleric's_wisdom_mod}` would tell the
battler what modifier to use for your *cure wounds* spell, as long as you set the `"modifier_key"` of that
spell's healing effect to be `"em1"` in the .json file. --dosav, --doatk, and --do are all just the
effects that you want to apply to the target you specified in the argument, how many times you want to
repeat those effects, and any advantage or disadvantage that needs to be considered for the rolls. 
Save-based effects, attack-based effects, and unavoidable effects are each stored in their own lists. The
first field of each of those flags is just the position of the desired effect in that list. In case 
nothing came to mind when I said "unavoidable effects", some examples would be the healing from
*cure wounds* and the damage from *magic missile*. The second field is the amount of times that effect
should be applied, like the amount of darts from *magic missile* or the amount of rays from *scorching ray*.
The next fields *can* be anything, but only `adv` and `dis` will be considered by the battler. `adv` means
advantage and `dis` means disadvantage. `--dosav 1 1 adv` means make the first save listed once with
advantage and take any necessary damage or receive and any necessary healing based on the result.
`--doatk 1 1 dis` means defened against the first spell attack listed once made at disadvantage and take any
necessary damage and receive any necessary healing based on the result. `--do 1 1` just means have the first
effect listed applied to you once, taking any necessary damage and receiving any necessary healing in the
process.
