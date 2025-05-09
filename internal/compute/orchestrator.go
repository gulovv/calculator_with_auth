package compute

import (
	"github.com/gulovv/calculator_with_auth/pkg/model"
	"fmt"
	"sync"
)

func Orchestrate(dagNodes []model.Node, known map[int]float64, finalID int, agentCount int) float64 {
    jobs := make(chan model.Node, len(dagNodes))
    results := make(chan model.Result, len(dagNodes))
    var knownMu sync.RWMutex

    for i := 1; i <= agentCount; i++ {
        go Agent(i, jobs, results, known, &knownMu)
    }

    for len(known) < finalID {
        var remaining []model.Node
        for _, node := range dagNodes {
            knownMu.RLock()
            _, hasLeft := known[node.LeftID]
            _, hasRight := known[node.RightID]
            knownMu.RUnlock()

            if hasLeft && hasRight {
				fmt.Printf("[Оркестратор] Отправляю на выполнение узел ID %d (%s)\n", node.ID, node.Op)
                jobs <- node
            } else {
                remaining = append(remaining, node)
            }
        }
        dagNodes = remaining



        res := <-results
		fmt.Printf("[Оркестратор] Получен результат: ID %d = %.2f\n", res.ID, res.Value)
        knownMu.Lock()
        known[res.ID] = res.Value
        knownMu.Unlock()
    }

    close(jobs)
    close(results)

    return known[finalID]
}