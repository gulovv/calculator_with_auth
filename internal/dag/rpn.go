package dag

import (
	"strconv"
	"fmt"
//	"strings"
)

func ToRPN(tokens []string) ([]string, error) {
	precedence := map[string]int{"+": 1, "-": 1, "*": 2, "/": 2}
	var output []string
	var stack []string

	for _, token := range tokens {
		if _, err := strconv.ParseFloat(token, 64); err == nil {
			output = append(output, token)
		} else if token == "(" {
			stack = append(stack, token)
		} else if token == ")" {
			foundLeftParen := false
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if top == "(" {
					foundLeftParen = true
					break
				}
				output = append(output, top)
			}
			if !foundLeftParen {
				return nil, fmt.Errorf("mismatched parentheses")
			}
		} else if _, ok := precedence[token]; ok {
			for len(stack) > 0 && precedence[stack[len(stack)-1]] >= precedence[token] {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, token)
		} else {
			return nil, fmt.Errorf("invalid token: %s", token)
		}
	}

	for len(stack) > 0 {
		top := stack[len(stack)-1]
		if top == "(" || top == ")" {
			return nil, fmt.Errorf("mismatched parentheses")
		}
		output = append(output, top)
		stack = stack[:len(stack)-1]
	}

	return output, nil
}