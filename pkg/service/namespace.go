package service

import "strings"

func namespaceForEnvironment(env string) string {
	value := strings.TrimSpace(strings.ToLower(env))
	value = strings.ReplaceAll(value, "_", "-")
	value = strings.ReplaceAll(value, " ", "-")
	if value == "" {
		return "default"
	}
	if strings.HasPrefix(value, "devflow-") {
		return value
	}
	return "devflow-" + value
}
