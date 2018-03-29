package main

import (
    "context"
    "net/http"
    //"time"
    "github.com/onrik/logrus/filename"
    log "github.com/sirupsen/logrus"
    "encoding/json"
    "TwitterClone/wrappers"
    "github.com/gorilla/mux"
    //"github.com/mongodb/mongo-go-driver/mongo"
    "github.com/mongodb/mongo-go-driver/bson"
)

type response struct {
    Status string `json:"status"`
    User User `json:"user,omitempty"`
    Error string `json:"error,omitempty"`
}

type User struct {
    Username string `json:"username bson:"username"`
    Email string `json:"email" bson:"email"`
    Password string `json:"password" bson:"password"`
    Followers int `json:"followers" bson:"followers"`
    Following int `json:"following" bson:"following"`
}

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/user/{username}", userHandler).Methods("GET")
    http.Handle("/", r)
    log.AddHook(filename.NewHook())
    log.SetLevel(log.DebugLevel)
    log.Fatal(http.ListenAndServe(":8007", nil))
}

func encodeResponse(w http.ResponseWriter, response interface{}) error {
    return json.NewEncoder(w).Encode(response)
}

func userHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    username := vars["username"]
    log.Debug(username)
    //var res response
    user, err := findUser(username)
    if err != nil {
        log.Info(err)
    }
    encodeResponse(w, user)
}

func findUser(username string) (*User, error) {
    client, err := wrappers.NewClient()
    if err != nil {
        return nil, err
    }
    db := client.Database("twitter")
    coll := db.Collection("users")
    filter := bson.NewDocument(bson.EC.String("username", username))
    result := bson.NewDocument()
    var user User
    err = coll.FindOne(context.Background(),
        filter).Decode(result)
    err = coll.FindOne(context.Background(),
        filter).Decode(&user)
    log.Debug(result)
    log.Debug(user)
    return nil, nil
}

