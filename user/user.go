package main

import (
    //"context"
    "net/http"
    //"time"
    "github.com/onrik/logrus/filename"
    log "github.com/sirupsen/logrus"
    //"encoding/json"
    "github.com/gorilla/mux"
    //"github.com/mongodb/mongo-go-driver/mongo"
    //"github.com/mongodb/mongo-go-driver/bson"
    //"github.com/mongodb/mongo-go-driver/bson/objectid"
)

type response struct {
    Status string `json:"status"`
    User User `json:"user,omitempty"`
    Error string `json:"error,omitempty"`
}

type User struct {
    Email string `json:"email"`
    Followers int `json:"followers"`
    Following int `json:"following"`

}

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/user/{username}", userHandler).Methods("GET")
    http.Handle("/", r)
    log.AddHook(filename.NewHook())
    log.SetLevel(log.DebugLevel)
    log.Fatal(http.ListenAndServe(":8007", nil))
}

func userHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    username := vars["username"]
    log.Info(username)
}

