package locales

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/user/kde-dotfiles-manager/internal/fileutil"
)

// LocaleConfig holds KDE language and locale configuration
type LocaleConfig struct {
	Language         string `yaml:"language"`
	Region           string `yaml:"region"`
	TimeFormat       string `yaml:"time_format"`
	DateFormat       string `yaml:"date_format"`
	NumberFormat     string `yaml:"number_format"`
	MonetaryFormat   string `yaml:"monetary_format"`
	Measurement      string `yaml:"measurement"`
	Collation        string `yaml:"collation"`
	InputMethod      string `yaml:"input_method"`
	KeyboardLayouts  []string `yaml:"keyboard_layouts"`
	SpellCheckLangs  []string `yaml:"spell_check_langs"`
}

// ParsePlasmaLocaleRc parses plasma-localerc for locale settings
func ParsePlasmaLocaleRc(path string) (*LocaleConfig, error) {
	cfg := &LocaleConfig{}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open plasma-localerc: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentSection := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line[1 : len(line)-1]
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch currentSection {
		case "Formats":
			switch key {
			case "LANG":
				cfg.Language = value
			case "LC_TIME":
				cfg.TimeFormat = value
			case "LC_MONETARY":
				cfg.MonetaryFormat = value
			case "LC_MEASUREMENT":
				cfg.Measurement = value
			case "LC_NUMERIC":
				cfg.NumberFormat = value
			case "LC_COLLATE":
				cfg.Collation = value
			}
		case "Translations":
			switch key {
			case "language":
				cfg.Language = value
			case "region":
				cfg.Region = value
			}
		}
	}

	return cfg, scanner.Err()
}

// ParseKdeglobalsLocale parses locale settings from kdeglobals
func ParseKdeglobalsLocale(path string) (map[string]string, error) {
	settings := make(map[string]string)

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open kdeglobals: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentSection := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line[1 : len(line)-1]
			continue
		}

		if currentSection == "Locale" {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				settings[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	return settings, scanner.Err()
}

// ParseFcitxConfig parses fcitx5 input method configuration
func ParseFcitxConfig(configDir string) (map[string]string, error) {
	settings := make(map[string]string)

	// Check common fcitx5 config locations
	profilePath := configDir + "/fcitx5/profile"
	if data, err := os.ReadFile(profilePath); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "Enabled Addons=") {
				settings["enabled_addons"] = strings.SplitN(line, "=", 2)[1]
			}
		}
	}

	return settings, nil
}

// Backup copies locale configuration files
func Backup(srcPath, destPath string) error {
	if !fileutil.FileExists(srcPath) {
		return nil
	}

	if err := fileutil.EnsureDir(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	return fileutil.CopyFile(srcPath, destPath)
}

// Restore copies locale configuration files back
func Restore(srcPath, destPath string) error {
	if !fileutil.FileExists(srcPath) {
		return nil
	}

	parent := filepath.Dir(destPath)
	if err := fileutil.EnsureDir(parent, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return fileutil.CopyFile(srcPath, destPath)
}
