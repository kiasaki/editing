package main

type Config struct {
	Colors   map[string]string
	Settings map[string]interface{}
}

func NewConfig() (*Config, error) {
	return &Config{
		Colors: map[string]string{
			"default":          "white",
			"comment":          "blue",
			"constant":         "red",
			"identifier":       "cyan",
			"statement":        "yellow",
			"preproc":          "magenta",
			"type":             "green",
			"special":          "magenta",
			"ignore":           "default",
			"error":            ",brightred",
			"todo":             ",brightyellow",
			"selection":        "black,brightyellow",
			"line-number":      "yellow",
			"gutter-info":      "blue",
			"gutter-error":     ",red",
			"gutter-warning":   "red",
			"statusbar":        "black,white",
			"statusbar-active": "black,brightwhite",
		},
		Settings: map[string]interface{}{
			"numbers":     true,
			"indent":      true,
			"tabtospaces": true,
			"tabwidth":    4,
			"light":       false,
		},
	}, nil
}

func (c *Config) GetSetting(name string) (interface{}, bool) {
	setting, ok := c.Settings[name]
	return setting, ok
}

func (c *Config) GetColor(name string) string {
	if color, ok := c.Colors[name]; ok {
		return color
	} else {
		return ""
	}
}
