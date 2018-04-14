package main

import (
    "net/http"
    "encoding/json"
    "os"
    "github.com/gorilla/mux"
    logrus "github.com/sirupsen/logrus"
    "TwitterClone/wrappers"
)


type response struct {
    Status string `json:"status"`
    Error string `json:"error,omitempty"`
}
var log *logrus.Logger
func main() {
    r := mux.NewRouter()
    r.HandleFunc("/logout", logoutHandler).Methods("POST")
    http.Handle("/", r)
    // Log to a file
    var f *os.File
    var err error
    log, f, err = wrappers.FileLogger("logout.log", os.O_CREATE | os.O_RDWR,
        0666)
    if err != nil {
        log.Fatal("Logging file could not be opened.")
    }
    f.Truncate(0)
    f.Seek(0, 0)
    defer f.Close()
    log.SetLevel(logrus.ErrorLevel)
    log.Fatal(http.ListenAndServe(":8001", nil))
}

func isLoggedIn(r *http.Request) bool {
    cookie, _ := r.Cookie("username")
    return cookie != nil
}

func encodeResponse(w http.ResponseWriter, response interface{}) error {
    return json.NewEncoder(w).Encode(response)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
    cookie, res := logoutEndpoint(r)
    if (cookie != nil) {
        http.SetCookie(w, cookie)
    }
    encodeResponse(w, res)
}

func logoutEndpoint(r *http.Request) (*http.Cookie, response) {
    var res response
    var cookie *http.Cookie
    if isLoggedIn(r) {
        cookie, _ = r.Cookie("username")
        cookie.MaxAge = -1 // Delete cookie.
        res.Status = "OK"
    } else {
        res.Status = "error"
        res.Error = "User not logged in."
    }
    return cookie, res
}
