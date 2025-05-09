package dag

import (
	"strconv"
	"github.com/gulovv/calculator_with_auth/pkg/model"
	"fmt"
)

func BuildDAG(rpn []string) (map[int]float64, []model.Node, int, error) {
	var idCounter = 1
	values := make(map[int]float64)
	var stack []int
	var nodes []model.Node

	for _, token := range rpn {
		if val, err := strconv.ParseFloat(token, 64); err == nil {
			values[idCounter] = val
			stack = append(stack, idCounter)
			idCounter++
		} else if isOperator(token) {
			// Нужно минимум два операнда в стеке
			if len(stack) < 2 {
				return nil, nil, 0, fmt.Errorf("not enough operands for operator %s", token)
			}

			right := stack[len(stack)-1]
			left := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			node := model.Node{
				ID:      idCounter,
				Op:      token,
				LeftID:  left,
				RightID: right,
			}

			nodes = append(nodes, node)
			stack = append(stack, idCounter)
			idCounter++
		} else {
			return nil, nil, 0, fmt.Errorf("invalid token in RPN: %s", token)
		}
	}

	if len(stack) != 1 {
		return nil, nil, 0, fmt.Errorf("invalid RPN: stack has %d elements, expected 1", len(stack))
	}

	finalID := stack[0]
	return values, nodes, finalID, nil
}

func isOperator(s string) bool {
	return s == "+" || s == "-" || s == "*" || s == "/"
}