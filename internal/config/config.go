package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type Schedule struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Pours []int  `json:"pours"`
}

type Tea struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Schedules []Schedule `json:"schedules"`
}

func (t *Tea) FindSchedule(id string) *Schedule {
	for i := range t.Schedules {
		if t.Schedules[i].ID == id {
			return &t.Schedules[i]
		}
	}
	return nil
}

type Config struct {
	Teas []Tea `json:"teas"`
}

func (c *Config) FindTea(id string) *Tea {
	for i := range c.Teas {
		if c.Teas[i].ID == id {
			return &c.Teas[i]
		}
	}
	return nil
}

var nonAlphaNum = regexp.MustCompile(`[^a-z0-9]+`)

func Slugify(s string) string {
	s = transliterate(s)
	s = strings.ToLower(s)
	s = nonAlphaNum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "tea"
	}
	return s
}

func transliterate(s string) string {
	transformer := transform.Chain(
		norm.NFD,
		transform.RemoveFunc(func(r rune) bool {
			return unicode.Is(unicode.Mn, r)
		}),
		norm.NFC,
	)
	result, _, _ := transform.String(transformer, s)

	cyrillic := map[rune]string{
		'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "yo",
		'ж': "zh", 'з': "z", 'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m",
		'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u",
		'ф': "f", 'х': "kh", 'ц': "ts", 'ч': "ch", 'ш': "sh", 'щ': "shch",
		'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
		'А': "A", 'Б': "B", 'В': "V", 'Г': "G", 'Д': "D", 'Е': "E", 'Ё': "Yo",
		'Ж': "Zh", 'З': "Z", 'И': "I", 'Й': "Y", 'К': "K", 'Л': "L", 'М': "M",
		'Н': "N", 'О': "O", 'П': "P", 'Р': "R", 'С': "S", 'Т': "T", 'У': "U",
		'Ф': "F", 'Х': "Kh", 'Ц': "Ts", 'Ч': "Ch", 'Ш': "Sh", 'Щ': "Shch",
		'Ъ': "", 'Ы': "Y", 'Ь': "", 'Э': "E", 'Ю': "Yu", 'Я': "Ya",
	}

	var b strings.Builder
	for _, r := range result {
		if mapped, ok := cyrillic[r]; ok {
			b.WriteString(mapped)
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func DefaultConfig() *Config {
	return &Config{
		Teas: []Tea{
			{
				ID:   "shen-puer",
				Name: "Шен Пуэр",
				Schedules: []Schedule{
					{ID: "default", Name: "Стандарт", Pours: []int{10, 15, 20, 25, 30, 40, 50, 60}},
					{ID: "fast", Name: "Быстрый", Pours: []int{5, 10, 10, 15, 15, 20}},
				},
			},
			{
				ID:   "shu-puer",
				Name: "Шу Пуэр",
				Schedules: []Schedule{
					{ID: "default", Name: "Стандарт", Pours: []int{10, 15, 15, 20, 25, 30, 40, 50}},
				},
			},
			{
				ID:   "da-khun-pao",
				Name: "Да Хун Пао",
				Schedules: []Schedule{
					{ID: "default", Name: "Стандарт", Pours: []int{15, 20, 25, 30, 35, 45, 60}},
				},
			},
		},
	}
}

func ConfigDir() string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "tmux-tea")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".config", "tmux-tea")
}

func DefaultPath() string {
	return filepath.Join(ConfigDir(), "teas.json")
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

func Save(cfg *Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

func LoadOrCreate(path string) (*Config, error) {
	cfg, err := Load(path)
	if err == nil {
		return cfg, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	cfg = DefaultConfig()
	if err := Save(cfg, path); err != nil {
		return nil, err
	}
	return cfg, nil
}
