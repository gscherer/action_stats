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
        wg.Add(3)
        jump_t := 10 + i
        walk_t := 12 + i
        run_t := 14 + i
        go func() {
            defer wg.Done()
            actions.AddAction(fmt.Sprintf(`{"action": "jump", "time": %d}`, jump_t))
            actions.GetStats()
        }()
        go func() {
            defer wg.Done()
            actions.AddAction(fmt.Sprintf(`{"action": "walk", "time": %d}`, walk_t))
            actions.GetStats()
        }()
        go func() {
            defer wg.Done()
            actions.AddAction(fmt.Sprintf(`{"action": "run", "time": %d}`, run_t))
            actions.GetStats()
        }()
    }
    wg.Wait()
    stats := actions.GetStats()
    var times [3]actionTime
    if err := json.Unmarshal([]byte(stats), &times); err != nil {
        t.Fatal(err)
    }
    jump_avg := 59;
    walk_avg := 61;
    run_avg := 63;
    seen := 0;
    for _, atm := range times {
        switch atm.Action {
        case "jump":
            if jump_avg != atm.Avg {
                t.Errorf("Expected avg of %d, got %d for action `%s`", jump_avg, atm.Avg, atm.Action)
            }
            seen += 1
        case "walk":
            if walk_avg != atm.Avg {
                t.Errorf("Expected avg of %d, got %d for action `%s`", walk_avg, atm.Avg, atm.Action)
            }
            seen += 1
        case "run":
            if run_avg != atm.Avg {
                t.Errorf("Expected avg of %d, got %d for action `%s`", run_avg, atm.Avg, atm.Action)
            }
            seen += 1
        }
    }
    if seen != 3 {
        t.Errorf("Result of GetStats did not include the expected entries")
    }
}
