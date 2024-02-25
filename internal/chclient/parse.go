package chclient

import (
	"regexp"
	"strings"
)

func MustParseSettings(engineFull string) map[string]string {
	settings := make(map[string]string)

	index := strings.Index(engineFull, "SETTINGS ")
	if index == -1 {
		return settings
	}

	input := strings.TrimSpace(engineFull[index+len("SETTINGS "):])

	for _, s := range strings.Split(input, ", ") {
		parts := strings.Split(s, " = ")
		if len(parts) != 2 {
			panic("Invalid settings string: " + engineFull)
		}
		key := parts[0]
		value := parts[1]
		value = strings.TrimLeft(value, "'")
		value = strings.TrimRight(value, "'")

		settings[key] = value
	}

	return settings
}

func MustParseEngineParams(engineFull string) []string {
	re := regexp.MustCompile(`\S\((.*?)\)`)
	matches := re.FindStringSubmatch(engineFull)
	if len(matches) == 0 {
		return make([]string, 0)
	}

	return strings.Split(matches[1], ", ")
}
