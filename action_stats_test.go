package action_stats

import (
    "fmt"
    "sync"
    "encoding/json"
    "testing"
)

func TestConcurrentActions(t *testing.T) {
    var wg sync.WaitGroup
    actions := NewActionMap()
    for i := 0; i < 100; i++ {
        t := 10 + i
        wg.Add(3)
        go func() {
            defer wg.Done()
            actions.AddAction(fmt.Sprintf(`{"action": "jump", "time": %d}`, t))
            actions.GetStats()
        }()
        go func() {
            defer wg.Done()
            actions.AddAction(fmt.Sprintf(`{"action": "walk", "time": %d}`, t))
            actions.GetStats()
        }()
        go func() {
            defer wg.Done()
            actions.AddAction(fmt.Sprintf(`{"action": "run", "time": %d}`, t))
            actions.GetStats()
        }()
    }
    wg.Wait()
    stats := actions.GetStats()
    var times [3]actionTime
    if err := json.Unmarshal([]byte(stats), &times); err != nil {
        t.Fatal(err)
    }
    for _, atm := range times {
        if atm.Avg == nil {
            t.Fatal("Average is nil")
        }
        if *atm.Avg != 59 {
            t.Errorf("Expected avg of 59, got %d for action `%s`", *atm.Avg, *atm.Action)
        }
    }
}
