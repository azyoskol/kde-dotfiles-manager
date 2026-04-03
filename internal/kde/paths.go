package kde

import "os"

// PlasmaVersion represents a KDE Plasma version
type PlasmaVersion string

const (
	Plasma6 PlasmaVersion = "6"
)

// Paths defines all KDE Plasma 6 configuration file locations
type Paths struct {
	// Config directory base
	ConfigDir string

	// Local directory base
	LocalDir string

	// Data directory base
	DataDir string
}

// NewPaths returns KDE paths for the current user
func NewPaths() (*Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := home + "/.config"
	localDir := home + "/.local"
	dataDir := home + "/.local/share"

	return &Paths{
		ConfigDir: configDir,
		LocalDir:  localDir,
		DataDir:   dataDir,
	}, nil
}

// KWinPaths returns paths for KWin (window manager) configuration
func (p *Paths) KWinPaths() map[string]string {
	return map[string]string{
		"kwinrc":                p.ConfigDir + "/kwinrc",
		"kwinrulesrc":           p.ConfigDir + "/kwinrulesrc",
		"kwin scripting":        p.DataDir + "/kwin/scripts",
		"kwin effects":          p.DataDir + "/kwin/effects",
		"kwin desktop effects":  p.ConfigDir + "/kwin-effects-rc",
		"virtual desktops":      p.ConfigDir + "/kdeglobals", // section [Desktops]
		"window behavior":       p.ConfigDir + "/kwinrc",
		"window rules":          p.ConfigDir + "/kwinrulesrc",
	}
}

// ShortcutPaths returns paths for keyboard shortcut configuration
func (p *Paths) ShortcutPaths() map[string]string {
	return map[string]string{
		"kglobalshortcutsrc": p.ConfigDir + "/kglobalshortcutsrc",
		"khotkeysrc":         p.ConfigDir + "/khotkeysrc",
	}
}

// ThemePaths returns paths for theme configuration
func (p *Paths) ThemePaths() map[string]string {
	return map[string]string{
		"kdeglobals":          p.ConfigDir + "/kdeglobals",
		"plasmarc":            p.ConfigDir + "/plasmarc",
		"auroraerc":           p.ConfigDir + "/auroraerc",
		"breezerc":            p.ConfigDir + "/breezerc",
		"kcminputrc":          p.ConfigDir + "/kcminputrc",
		"gtkrc":               p.ConfigDir + "/gtkrc",
		"gtk-3.0 settings":    p.ConfigDir + "/gtk-3.0/settings.ini",
		"gtk-4.0 settings":    p.ConfigDir + "/gtk-4.0/settings.ini",
		"color schemes":       p.DataDir + "/color-schemes",
		"wallpapers":          p.DataDir + "/wallpapers",
		"icons":               p.DataDir + "/icons",
		"fonts":               p.DataDir + "/fonts",
		"sounds":              p.DataDir + "/sounds",
		"look and feel":       p.DataDir + "/plasma/look-and-feel",
		"window decorations":  p.DataDir + "/aurorae/themes",
		"cursor themes":       p.DataDir + "/icons",
	}
}

// LocalePaths returns paths for language/locale configuration
func (p *Paths) LocalePaths() map[string]string {
	return map[string]string{
		"kdeglobals locale":   p.ConfigDir + "/kdeglobals",
		"plasma-localerc":     p.ConfigDir + "/plasma-localerc",
		"language config":     p.ConfigDir + "/language.conf",
		"input method config": p.ConfigDir + "/im-config.conf",
		"ibus config":         p.ConfigDir + "/ibus",
		"fcitx config":        p.ConfigDir + "/fcitx",
		"fcitx5 config":       p.ConfigDir + "/fcitx5",
	}
}

// WidgetPaths returns paths for widget configuration
func (p *Paths) WidgetPaths() map[string]string {
	return map[string]string{
		"plasma widgets":     p.DataDir + "/plasma/plasmoids",
		"plasma layout":      p.DataDir + "/plasma/layout-templates",
		"plasma packages":    p.DataDir + "/plasma/packages",
		"desktop containment": p.DataDir + "/plasma/org.kde.plasma.desktop-appletsrc",
	}
}

// PanelPaths returns paths for panel configuration
func (p *Paths) PanelPaths() map[string]string {
	return map[string]string{
		"panel layout": p.DataDir + "/plasma/org.kde.panel",
		"plasmarc":     p.ConfigDir + "/plasmarc",
	}
}

// SystemSettingsPaths returns paths for general system settings
func (p *Paths) SystemSettingsPaths() map[string]string {
	return map[string]string{
		"kdeglobals":        p.ConfigDir + "/kdeglobals",
		"systemsettingsrc":  p.ConfigDir + "/systemsettingsrc",
		"powerdevilrc":      p.ConfigDir + "/powerdevilrc",
		"kscreenlockerrc":   p.ConfigDir + "/kscreenlockerrc",
		"kded5rc":           p.ConfigDir + "/kded5rc",
		"kded6rc":           p.ConfigDir + "/kded6rc",
		"ksplashrc":         p.ConfigDir + "/ksplashrc",
		"startkderc":        p.ConfigDir + "/startkderc",
		"ksmserverrc":       p.ConfigDir + "/ksmserverrc",
		"krunnerrc":         p.ConfigDir + "/krunnerrc",
		"kwalletrc":         p.ConfigDir + "/kwalletrc",
		"baloofilerc":       p.ConfigDir + "/baloofilerc",
		"dolphinrc":         p.ConfigDir + "/dolphinrc",
		"katerc":            p.ConfigDir + "/katerc",
		"konsoleshellrc":    p.ConfigDir + "/konsoleshellrc",
	}
}

// AllConfigFiles returns a map of all KDE configuration files
func (p *Paths) AllConfigFiles() map[string]string {
	all := make(map[string]string)
	for k, v := range p.KWinPaths() {
		all["kwin."+k] = v
	}
	for k, v := range p.ShortcutPaths() {
		all["shortcuts."+k] = v
	}
	for k, v := range p.ThemePaths() {
		all["themes."+k] = v
	}
	for k, v := range p.LocalePaths() {
		all["locales."+k] = v
	}
	for k, v := range p.WidgetPaths() {
		all["widgets."+k] = v
	}
	for k, v := range p.PanelPaths() {
		all["panels."+k] = v
	}
	for k, v := range p.SystemSettingsPaths() {
		all["system."+k] = v
	}
	return all
}
