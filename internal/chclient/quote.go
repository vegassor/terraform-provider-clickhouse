// Inspired by: https://github.com/go-rel/postgres/blob/main/quote.go

package chclient

import (
	"strings"
)

// QuoteID quotes ClickHouse identifiers in order to try to prevent SQL injection.
func QuoteID(name string) string {
	end := strings.IndexRune(name, 0)
	if end > -1 {
		name = name[:end]
	}
	name = strings.ReplaceAll(name, `\`, `\\`)
	name = strings.ReplaceAll(name, `"`, `\"`)

	return `"` + name + `"`
}

// QuoteValue quotes ClickHouse literals in order to try to prevent SQL injection.
func QuoteValue(v string) string {
	end := strings.IndexRune(v, 0)
	if end > -1 {
		v = v[:end]
	}
	v = strings.ReplaceAll(v, `\`, `\\`)
	v = strings.ReplaceAll(v, `'`, `\'`)

	return `'` + v + `'`
}

// QuoteWithTicks quotes ClickHouse identifiers with ticks (` `) in order to try to prevent SQL injection.
func QuoteWithTicks(v string) string {
	end := strings.IndexRune(v, 0)
	if end > -1 {
		v = v[:end]
	}
	v = strings.ReplaceAll(v, `\`, `\\`)
	v = strings.ReplaceAll(v, "`", "\\`")

	return "`" + v + "`"
}

func QuoteList(v []string, quote string) []string {
	result := make([]string, 0, len(v))

	for _, s := range v {
		val := s
		end := strings.IndexRune(val, 0)
		if end > -1 {
			val = val[:end]
		}
		val = strings.ReplaceAll(val, `\`, `\\`)
		val = strings.ReplaceAll(val, quote, "\\"+quote)
		result = append(result, quote+val+quote)
	}

	return result
}

func QuoteListWithTicksAndJoin(v []string) string {
	result := QuoteList(v, "`")
	return strings.Join(result, ", ")
}

func QuoteMapAndJoin(data map[string]string) string {
	var params []string

	for k, v := range data {
		key := QuoteWithTicks(k)
		val := QuoteValue(v)
		params = append(params, key+" = "+val)
	}

	return strings.Join(params, ", ")
}
