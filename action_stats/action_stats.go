package action_stats

import (
    "sync"
    "encoding/json"
)

type actionTime struct {
    Action string   `json:"action"`
    Time   int      `json:"time,omitempty"`
    Avg    int      `json:"avg,omitempty"`
}

type actionStat struct {
    count int
    totalTime int
}

func (s *actionStat) avg() int {
    return s.totalTime / s.count
}

func (s *actionStat) addTime(time int) {
    s.count++
    s.totalTime += time
}

type actionMap struct {
    sync.RWMutex
    store map[string]*actionStat
}

func (a *actionMap) addActionTime(atm actionTime) {
    stat, ok := a.store[atm.Action]
    if !ok {
        a.store[atm.Action] = &actionStat{count: 1, totalTime: atm.Time}
    } else {
        stat.addTime(atm.Time)
    }
}

func (a *actionMap) MarshalJSON() ([]byte, error) {
    a.RLock()
    defer a.RUnlock()
    res := make([]actionTime, len(a.store))
    i := 0
    for action, stat := range a.store {
        res[i] = actionTime{Action: action, Avg: stat.avg()}
        i++
    }
    return json.Marshal(res)
}

func (a *actionMap) AddAction(req string) error {
    a.Lock()
    defer a.Unlock()
    var atm actionTime
    if err := json.Unmarshal([]byte(req), &atm); err != nil {
        return err
    }
    a.addActionTime(atm)
    return nil
}

func (a *actionMap) GetStats() string {
    res, _ := a.MarshalJSON()
    return string(res)
}

func NewActionMap() *actionMap {
    var a actionMap
    a.store = make(map[string]*actionStat)
    return &a
}
