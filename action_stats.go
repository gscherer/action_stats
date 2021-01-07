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

// The actionTime type is what represents both the requests to add an "action time"
// and the responses which show the average time for each action. This struct is
// used to unmarshal JSON requests to AddAction and when marshalling json responses 
// for GetStats.
type actionTime struct {
    Action string   `json:"action"`
    Time   int      `json:"time,omitempty"`
    Avg    int      `json:"avg,omitempty"`
}

// The actionStat type is used to store the current state of an action, both
// how many times it has been added, and the total time used for that action.
// It has two methods, avg() and addTime(t int) to change its state in response
// to requests to AddAction
type actionStat struct {
    count int
    totalTime int
}

// Get the current average time for a particular action, based on all requests
// made to AddAction for that action.
func (s *actionStat) avg() int {
    return s.totalTime / s.count
}

// Add time to a particular action and increment the count of requests made to
// do so. This is called for each valid request to AddAction.
func (s *actionStat) addTime(time int) {
    s.count++
    s.totalTime += time
}

// The actionMap type stores a map of actions to actionStat, keeping track
// of the state of each action. It is protected by concurrent reads/writes by
// a read/write mutex.
type actionMap struct {
    sync.RWMutex
    store map[string]*actionStat
}

// From a request to add time to an action, this method will validate
// that the requested action and time are valid and either add the action
// to the map, or increment the count and total time for the action if 
// it already exists.
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

// Implements the MarshalJSON interface so that the actionMap type can
// be passed json.MarshalJSON(). This is used in the GetStats method,
// to return the entire state of the actionMap in JSON format.
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

// Public method used to add action and time for the action.
// The request is a JSON string in the following format:
//      { "action": "someaction", "time": 10 }
// The "time" field must be an integer > 0.
// This method may be called safely concurrently.
func (a *actionMap) AddAction(req string) error {
    a.Lock()
    defer a.Unlock()
    var atm actionTime
    if err := json.Unmarshal([]byte(req), &atm); err != nil {
        return err
    }
    return a.addActionTime(atm)
}

// Public method used to return the entire state of the actionMap as
// a JSON string. It will return an array of object where each object
// holds the action name and current average time for that action.
// This method may be called safely concurrently.
func (a *actionMap) GetStats() string {
    res, _ := a.MarshalJSON()
    return string(res)
}

// Get a new instance of an actionMap. It won't hold any values until
// a request to AddAction is made.
func NewActionMap() *actionMap {
    var a actionMap
    a.store = make(map[string]*actionStat)
    return &a
}

// Request handler for calls to GET /action-stats. Returns an HTTP response of
// content type "application/json" with the results of GetStats.
func (a *actionMap) handleHttpGet(res http.ResponseWriter, req *http.Request) {
    res.Header().Set("Content-Type", "application/json; charset=UTF-8")
    res.WriteHeader(http.StatusOK)
    fmt.Fprintf(res, a.GetStats())
}

// Request handler for calls to POST /action-stats. Returns an HTTP response code of 201 
// if the call to AddAction was successful, or 400 if the request was invalid. The body of the 
// request should be the same JSON string used as the argument to AddAction.
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

// Used to represent error states in the HTTP server
type jsonError struct {
    Error   string   `json:"error"`
    Message string   `json:"message"`
}

// Send a JSON formatted body for error responses.
func httpJsonError(res http.ResponseWriter, message string, code int) {
    res.Header().Set("Content-Type", "application/json; charset=UTF-8")
    err := jsonError{Error: http.StatusText(code), Message: message}
    body, _ := json.Marshal(err) // TODO(grayson) handle encoding issue with 500
    res.WriteHeader(code)
    fmt.Fprintf(res, string(body))
}

// Public function used to start the HTTP server. This server hosts two endpoints:
//      GET /action-stats
//      POST /action-stats
// When started, it creates a new actionMap, which is modified by successful requests to POST /action-stats.
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
