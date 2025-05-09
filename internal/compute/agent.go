package compute

import (
	"github.com/gulovv/calculator_with_auth/pkg/model"
	"fmt"
	"sync"
	"time"
)

func Agent(id int, jobs chan model.Node, results chan model.Result, known map[int]float64, knownMu *sync.RWMutex) {
	for node := range jobs {
		knownMu.RLock()
		left := known[node.LeftID]
		right := known[node.RightID]
		knownMu.RUnlock()

		var result float64
		switch node.Op {
		case "+":
			result = left + right
		case "-":
			result = left - right
		case "*":
			result = left * right
		case "/":
			result = left / right
		}
		fmt.Printf("[Агент %d] Выполняю: %.2f %s %.2f = %.2f (ID %d)\n",
		id, left, node.Op, right, result, node.ID)


		time.Sleep(200 * time.Millisecond) // Имитация вычислений
		results <- model.Result{ID: node.ID, Value: result}
	}
}