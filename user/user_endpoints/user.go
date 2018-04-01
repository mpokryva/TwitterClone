package main

import (
    "context"
    "net/http"
    "github.com/onrik/logrus/filename"
    log "github.com/sirupsen/logrus"
    "encoding/json"
    "TwitterClone/wrappers"
    "github.com/gorilla/mux"
    "github.com/mongodb/mongo-go-driver/bson"
    "TwitterClone/user"
)

type response struct {
    Status string `json:"status"`
    User *user.User `json:"user,omitempty"`
    Error string `json:"error,omitempty"`
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
    var res response
    user, err := findUser(username)
    if err != nil {
        log.Info(err)
        res.Status = "error"
        res.Error = err.Error()
    } else {
        log.Debug(user)
        res.Status = "OK"
        user.Password = ""
        user.Key = ""
        user.Verified = false
        user.Email = ""
        res.User = user
    }
    encodeResponse(w, res)
}

func findUser(username string) (*user.User, error) {
    client, err := wrappers.NewClient()
    if err != nil {
        return nil, err
    }
    db := client.Database("twitter")
    coll := db.Collection("users")
    filter := bson.NewDocument(bson.EC.String("username", username))
    var user user.User
    err = coll.FindOne(context.Background(),
        filter).Decode(&user)
    return &user, err
}

