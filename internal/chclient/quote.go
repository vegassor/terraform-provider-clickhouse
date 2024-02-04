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
	v = strings.ReplaceAll(v, `'`, `\'`)

	return "`" + v + "`"
}
