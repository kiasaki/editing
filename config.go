package main

var (
	config map[string]interface{}
)

func init_config() {
	config = map[string]interface{}{
		"tab_width":     float64(4),
		"tab_to_spaces": true,
	}
}

func config_get(key string, b *buffer) string {
	if v, ok := config[key]; ok {
		if vv, ok := v.(string); ok {
			return vv
		}
	}
	return ""
}

func config_get_bool(key string, b *buffer) bool {
	if v, ok := config[key]; ok {
		if vv, ok := v.(bool); ok {
			return vv
		}
	}
	return false
}

func config_get_number(key string, b *buffer) float64 {
	if v, ok := config[key]; ok {
		if vv, ok := v.(float64); ok {
			return vv
		}
	}
	return 0
}

func config_set(key string, value interface{}) {
	config[key] = value
}
