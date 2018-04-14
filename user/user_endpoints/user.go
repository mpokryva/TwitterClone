package main

import (
    "context"
    "net/http"
    logrus "github.com/sirupsen/logrus"
    "os"
    "encoding/json"
    "TwitterClone/wrappers"
    "github.com/gorilla/mux"
    "github.com/mongodb/mongo-go-driver/bson"
    "TwitterClone/user"
)

type response struct {
    Status string `json:"status"`
    User userResponse `json:"user,omitempty"`
    Error string `json:"error,omitempty"`
}

type userResponse struct {
    Email string `json:"email"`
    Followers int `json:"followers"`
    Following int `json:"following"`
}
var log *logrus.Logger
func main() {
    r := mux.NewRouter()
    r.HandleFunc("/user/{username}", userHandler).Methods("GET")
    http.Handle("/", r)
    // Log to a file
    var f *os.File
    var err error
    log, f, err = wrappers.FileLogger("user.log", os.O_CREATE | os.O_RDWR,
        0666)
    if err != nil {
        log.Fatal("Logging file could not be opened.")
    }
    f.Truncate(0)
    f.Seek(0, 0)
    defer f.Close()
    log.SetLevel(logrus.ErrorLevel)
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
        var userRes userResponse
        userRes.Email = user.Email
        userRes.Following = user.FollowingCount
        userRes.Followers = user.FollowerCount
        res.User = userRes
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

