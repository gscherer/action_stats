package action_stats

import (
    "fmt"
    "errors"
    "log"
    "io/ioutil"
    "sync"
    "encoding/json"
    "net/http"
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

func (a *actionMap) addActionTime(atm actionTime) error {
    if atm.Action == "" || atm.Time <= 0 {
        return errors.New("Invalid action or time")
    }
    action := atm.Action
    time := atm.Time
    stat, ok := a.store[action]
    if !ok {
        a.store[action] = &actionStat{count: 1, totalTime: time}
    } else {
        stat.addTime(time)
    }
    return nil
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
    return a.addActionTime(atm)
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

func (a *actionMap) handleHttpGet(res http.ResponseWriter, req *http.Request) {
    res.Header().Set("Content-Type", "application/json; charset=UTF-8")
    res.WriteHeader(http.StatusOK)
    fmt.Fprintf(res, a.GetStats())
}

func (a *actionMap) handleHttpPost(res http.ResponseWriter, req *http.Request) {
    body, err := ioutil.ReadAll(req.Body) // Probably a better way to do this? Body could be huge...
    if err != nil {
        log.Println(err)
        httpJsonError(res, "Failed to read request body", http.StatusInternalServerError)
        return
    }
    err = a.AddAction(string(body))
    if err != nil {
        log.Println(err)
        httpJsonError(res, "Invalid action or time", http.StatusBadRequest)
        return
    }
    res.Header().Set("Content-Type", "application/json; charset=UTF-8")
    res.WriteHeader(http.StatusCreated)
    fmt.Fprintf(res, a.GetStats())
}

type jsonError struct {
    Error   string   `json:"error"`
    Message string   `json:"message"`
}

func httpJsonError(res http.ResponseWriter, message string, code int) {
    res.Header().Set("Content-Type", "application/json; charset=UTF-8")
    err := jsonError{Error: http.StatusText(code), Message: message}
    body, _ := json.Marshal(err) // TODO(grayson) handle encoding issue with 500
    res.WriteHeader(code)
    fmt.Fprintf(res, string(body))
}

func StartServer(port string) {
    actionMap := NewActionMap()
    mux := http.NewServeMux()
    mux.HandleFunc("/action-stats", func (res http.ResponseWriter, req *http.Request) {
        switch req.Method {
            case http.MethodGet:
                actionMap.handleHttpGet(res, req)
            case http.MethodPost:
                actionMap.handleHttpPost(res, req)
            default:
                httpJsonError(res, "", http.StatusNotFound)
        }
    })
    err := http.ListenAndServe(port, mux)
    if err != nil {
        log.Fatal(err)
    }
}
