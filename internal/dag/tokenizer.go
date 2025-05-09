package dag

import (
	"strings"
	"fmt"
)

func Tokenize(expr string) ([]string, error) {
	var tokens []string
	var current strings.Builder
	expr = strings.ReplaceAll(expr, " ", "")
	prevTokenType := "op"

	for i := 0; i < len(expr); i++ {
		ch := expr[i]

		// Число или унарный минус
		if (ch >= '0' && ch <= '9') || ch == '.' || (ch == '-' && prevTokenType == "op") {
			current.WriteByte(ch)
			dotCount := 0
			if ch == '.' {
				dotCount++
			}

			for i+1 < len(expr) && ((expr[i+1] >= '0' && expr[i+1] <= '9') || expr[i+1] == '.') {
				i++
				if expr[i] == '.' {
					dotCount++
					if dotCount > 1 {
						return nil, fmt.Errorf("invalid number with multiple dots")
					}
				}
				current.WriteByte(expr[i])
			}
			tokens = append(tokens, current.String())
			current.Reset()
			prevTokenType = "num"
		} else if strings.Contains("+-*/()", string(ch)) {
			tokens = append(tokens, string(ch))
			prevTokenType = "op"
		} else {
			return nil, fmt.Errorf("invalid character: %c", ch)
		}
	}

	fmt.Println("Токены:", tokens)
	return tokens, nil
}