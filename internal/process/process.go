package process

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/45uperman/dndbattlercli/internal/battler"
	"github.com/45uperman/dndbattlercli/internal/battler/combatant"
)

func SaveFiles(b battler.Battler) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error finding program path: %w", err)
	}

	root := filepath.Dir(exePath)
	absPath := root + "/battle_files/"

	b.MU.RLock()
	defer b.MU.RUnlock()

	returnAddresses := make(map[string][]byte, len(b.Combatants))
	for _, combatant := range b.Combatants {
		returnAddresses[combatant.StatBlock.FileName], err = json.MarshalIndent(combatant, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshaling %s: %s", combatant.StatBlock.Name, err)
		}
	}

	for address, data := range returnAddresses {
		err = os.WriteFile(absPath+address, data, 0777)
		if err != nil {
			fmt.Printf("error writing file %s: %s\n", address, err)
		}
	}

	return nil
}

func LoadFiles() (battler.Battler, error) {
	exePath, err := os.Executable()
	if err != nil {
		return battler.Battler{}, fmt.Errorf("error finding program path: %w", err)
	}

	root := filepath.Dir(exePath)
	absPath := root + "/battle_files"

	newBattler := battler.NewBattler()
	err = filepath.Walk(
		absPath,
		func(path string, info os.FileInfo, err error) error {
			return loadFile(path, info, &newBattler, err)
		},
	)
	if err != nil {
		return battler.Battler{}, fmt.Errorf("could not load files because of error: %s", err)
	}

	return newBattler, nil
}

func loadFile(path string, info os.FileInfo, b *battler.Battler, err error) error {
	if err != nil {
		return err
	}

	if filepath.Ext(path) == ".json" {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var c combatant.Combatant
		err = json.Unmarshal(data, &c)
		if err != nil {
			return err
		}

		b.AddCombatant(c)
	}
	return nil
}
