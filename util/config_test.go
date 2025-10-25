package util

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/charmbracelet/bubbles/key"
)

func TestKeyMapEncoding(t *testing.T) {
	orig := DefaultKeyMap()

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Could not marshal keymap: %s", err.Error())
	}
	jsonStr := string(data)

	var decoded KeyMap
	if err := json.Unmarshal([]byte(jsonStr), &decoded); err != nil {
		t.Fatalf("Could not unmarshal keymap: %v", err)
	}

	if !reflect.DeepEqual(orig, decoded) {
		t.Fatalf("Enc/Dec is not bijective\nOriginal: %+v\nDecoded:  %+v", orig, decoded)
	}
}

func TestPartialKeyMapDecoding(t *testing.T) {
	partial := `{"up":{"keys":["up"]}}`

	var km KeyMap
	if err := json.Unmarshal([]byte(partial), &km); err != nil {
		t.Fatalf("Could not unmarshal partial keymap: %v", err)
	}

	if len(km.Down.Keys()) != 0 {
		t.Errorf("Expected Down to be empty, got keys: %v", km.Down.Keys())
	}

	if len(km.Up.Keys()) != 1 || km.Up.Keys()[0] != "up" {
		t.Errorf("Up did not decode correctly. Expected: %v, Got: %v", km.Up, key.NewBinding(key.WithKeys("up")))
	}
}

func TestCreatesMissingConfig(t *testing.T) {
	testDir := os.TempDir
	testCfgPath := configPath(testDir())
	if _, err := os.Stat(testCfgPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Config already exists: %v", err)
	}
	_, err := LoadConfig(testDir())
	if err != nil {
		t.Fatalf("Loading config failed: %v", err)
	}
	if _, err := os.Stat(testCfgPath); errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Did not create a config file: %v", err)
	}
	if err := os.Remove(testCfgPath); err != nil {
		t.Fatalf("Could not remove the test config: %v", err)
	}
}
