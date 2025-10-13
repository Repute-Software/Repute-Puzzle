package models

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// Translations holds all translated strings for the application
type Translations struct {
	Game struct {
		Title        string `yaml:"title"`
		Subtitle     string `yaml:"subtitle"`
		Moves        string `yaml:"moves"`
		Time         string `yaml:"time"`
		NewPuzzle    string `yaml:"new_puzzle"`
		ShowSolution string `yaml:"show_solution"`
		AutoSolve    string `yaml:"auto_solve"`
		Solving      string `yaml:"solving"`
	} `yaml:"game"`
	Modal struct {
		Congratulations  string `yaml:"congratulations"`
		Completed        string `yaml:"completed"`
		MovesLabel       string `yaml:"moves_label"`
		TimeLabel        string `yaml:"time_label"`
		EmailPrompt      string `yaml:"email_prompt"`
		EmailPlaceholder string `yaml:"email_placeholder"`
		GetCodeButton    string `yaml:"get_code_button"`
	} `yaml:"modal"`
	Completion struct {
		Title       string `yaml:"title"`
		SaveInfo    string `yaml:"save_info"`
		Instruction string `yaml:"instruction"`
		EmailSent   string `yaml:"email_sent"`
		PlayAgain   string `yaml:"play_again"`
	} `yaml:"completion"`
	Error struct {
		Title          string `yaml:"title"`
		TryAgain       string `yaml:"try_again"`
		InvalidEmail   string `yaml:"invalid_email"`
		FailedGenerate string `yaml:"failed_generate"`
		NoImages       string `yaml:"no_images"`
	} `yaml:"error"`
}

// Translation cache to prevent file descriptor exhaustion
var (
	translationCache = make(map[string]*Translations)
	cacheMutex       sync.RWMutex
)

// InitTranslationCache pre-loads all translations into memory
func InitTranslationCache() error {
	languages := []string{"en", "nl"}
	
	for _, lang := range languages {
		translations, err := loadTranslationFile(lang)
		if err != nil {
			return fmt.Errorf("failed to load %s translations: %w", lang, err)
		}
		
		cacheMutex.Lock()
		translationCache[lang] = translations
		cacheMutex.Unlock()
	}
	
	return nil
}

// loadTranslationFile loads a translation file from disk
func loadTranslationFile(lang string) (*Translations, error) {
	filename := fmt.Sprintf("locales/%s.yaml", lang)
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read translation file %s: %w", filename, err)
	}

	var translations Translations
	if err := yaml.Unmarshal(data, &translations); err != nil {
		return nil, fmt.Errorf("failed to parse translation file %s: %w", filename, err)
	}

	return &translations, nil
}

// LoadTranslations loads translation strings for the specified language from cache
func LoadTranslations(lang string) (*Translations, error) {
	// Default to English
	if lang == "" {
		lang = "en"
	}

	// Validate language (only en and nl supported)
	if lang != "en" && lang != "nl" {
		lang = "en"
	}

	// Try to get from cache first
	cacheMutex.RLock()
	translations, ok := translationCache[lang]
	cacheMutex.RUnlock()

	if ok {
		return translations, nil
	}

	// If not in cache (shouldn't happen after InitTranslationCache), load from file
	translations, err := loadTranslationFile(lang)
	if err != nil {
		return nil, err
	}

	// Store in cache
	cacheMutex.Lock()
	translationCache[lang] = translations
	cacheMutex.Unlock()

	return translations, nil
}
