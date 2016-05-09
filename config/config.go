package config

type Config struct {
	colors   map[string]string
	settings map[string]interface{}
}

func ConfigNew() *Config {
	return &Config{
		colors: map[string]string{
			"comment":        "blue",
			"constant":       "red",
			"identifier":     "cyan",
			"statement":      "yellow",
			"preproc":        "magenta",
			"type":           "green",
			"special":        "magenta",
			"ignore":         "default",
			"error":          ",brightred",
			"todo":           ",brightyellow",
			"selection":      "black,brightyellow",
			"line-number":    "yellow",
			"gutter-info":    "blue",
			"gutter-error":   ",red",
			"gutter-warning": "red",
			"statusbar":      "black,white",
		},
		settings: map[string]interface{}{
			"numbers":     true,
			"indent":      true,
			"tabtospaces": true,
			"tabwidth":    4,
			"light":       false,
		},
	}
}

func (c *Config) GetSetting(name string) (interface{}, bool) {
	setting, ok := c.settings[name]
	return setting, ok
}

func (c *Config) GetColor(name string) string {
	if color, ok := c.colors[name]; ok {
		return color
	} else {
		return ""
	}
}
