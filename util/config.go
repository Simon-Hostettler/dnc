package util

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/charmbracelet/bubbles/key"
)

func DefaultConfigDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		panic("Cannot start without a writable data directory.")
	}
	return configDir
}

func configPath(cfgDir string) string {
	return filepath.Join(dncConfigDir(cfgDir), "config.json")
}

func dncConfigDir(cfgDir string) string {
	return filepath.Join(cfgDir, "dnc")
}

func encode(f *os.File, v any) error {
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

type Config struct {
	KeyMap       KeyMap `json:"keymap"`
	DatabasePath string `json:"database_path"`
}

func DefaultConfig(cfgDir string) Config {
	return Config{
		KeyMap:       DefaultKeyMap(),
		DatabasePath: filepath.Join(cfgDir, "dnc", "dnc.db"),
	}
}

func CreateConfigIfMissing(cfgDir string) error {
	cfgPath := configPath(cfgDir)
	if _, err := os.Stat(cfgPath); err == nil {
		return nil // config exists
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(dncConfigDir(cfgDir), 0o755); err != nil {
		return err
	}
	f, err := os.Create(cfgPath)
	if err != nil {
		return err
	}
	defer f.Close()
	cfg := DefaultConfig(cfgDir)
	if err := encode(f, cfg); err != nil {
		return err
	}
	return nil
}

func LoadConfig(cfgDir string) (Config, error) {
	def := DefaultConfig(cfgDir)
	err := CreateConfigIfMissing(cfgDir)
	if err != nil {
		return def, err
	}
	cfgPath := configPath(cfgDir)
	f, err := os.Open(cfgPath)
	if err != nil {
		return def, nil
	}
	defer f.Close()
	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return def, err
	}
	if cfg.DatabasePath == "" {
		cfg.DatabasePath = def.DatabasePath
	}
	return cfg, nil
}

type KeyMap struct {
	Up        key.Binding `json:"up"`
	Down      key.Binding `json:"down"`
	Left      key.Binding `json:"left"`
	Right     key.Binding `json:"right"`
	Select    key.Binding `json:"select"`
	Edit      key.Binding `json:"edit"`
	Enter     key.Binding `json:"enter"`
	Escape    key.Binding `json:"escape"`
	Delete    key.Binding `json:"delete"`
	ForceQuit key.Binding `json:"force_quit"`
	Show      key.Binding `json:"show"`
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up:        key.NewBinding(key.WithKeys("up")),
		Down:      key.NewBinding(key.WithKeys("down")),
		Left:      key.NewBinding(key.WithKeys("left")),
		Right:     key.NewBinding(key.WithKeys("right")),
		Select:    key.NewBinding(key.WithKeys(" ", "enter")),
		Edit:      key.NewBinding(key.WithKeys("e")),
		Enter:     key.NewBinding(key.WithKeys("enter")),
		Escape:    key.NewBinding(key.WithKeys("esc", "q")),
		Delete:    key.NewBinding(key.WithKeys("x", "del")),
		ForceQuit: key.NewBinding(key.WithKeys("ctrl+c")),
		Show:      key.NewBinding(key.WithKeys(" ")),
	}
}

type keyBindingDTO struct {
	Keys []string `json:"keys,omitempty"`
}

func toDTO(b key.Binding) keyBindingDTO {
	return keyBindingDTO{Keys: b.Keys()}
}

func fromDTO(d keyBindingDTO) key.Binding {
	if len(d.Keys) == 0 {
		return key.Binding{}
	}
	return key.NewBinding(key.WithKeys(d.Keys...))
}

// custom Marshallers as key.Binding cannot be encoded
func (km KeyMap) MarshalJSON() ([]byte, error) {
	m := make(map[string]keyBindingDTO)

	kmVal := reflect.ValueOf(km)
	kmType := kmVal.Type()
	for i := 0; i < kmType.NumField(); i++ {
		kmField := kmType.Field(i)
		if kmField.Type != reflect.TypeOf(key.Binding{}) {
			return []byte{}, errors.New("a keymap should only contain fields of type key.Binding")
		}
		tag := kmField.Tag.Get("json")
		name := strings.Split(tag, ",")[0]
		if name == "" || name == "-" {
			continue
		}
		b := kmVal.Field(i).Interface().(key.Binding)
		m[name] = toDTO(b)
	}
	return json.Marshal(m)
}

func (km *KeyMap) UnmarshalJSON(data []byte) error {
	var m map[string]keyBindingDTO
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	kmVal := reflect.ValueOf(km).Elem()
	kmType := kmVal.Type()
	for i := 0; i < kmType.NumField(); i++ {
		kmField := kmType.Field(i)
		if kmField.Type != reflect.TypeOf(key.Binding{}) {
			return errors.New("a keymap should only contain fields of type key.Binding")
		}
		tag := kmField.Tag.Get("json")
		name := strings.Split(tag, ",")[0]
		if name == "" || name == "-" {
			continue
		}
		if dto, ok := m[name]; ok {
			kmVal.Field(i).Set(reflect.ValueOf(fromDTO(dto)))
		} else {
			kmVal.Field(i).Set(reflect.ValueOf(key.Binding{}))
		}
	}
	return nil
}
