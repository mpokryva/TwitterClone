package user_endpoints

import (
    "context"
    "net/http"
    logrus "github.com/sirupsen/logrus"
    
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
var Log *logrus.Logger
func main() {
    r := mux.NewRouter()
    r.HandleFunc("/user/{username}", GetUserHandler).Methods("GET")
    http.Handle("/", r)
    Log.SetLevel(logrus.ErrorLevel)
}

func encodeResponse(w http.ResponseWriter, response interface{}) error {
    return json.NewEncoder(w).Encode(response)
}

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    username := vars["username"]
    Log.Debug(username)
    var res response
    user, err := findUser(username)
    if err != nil {
        Log.Info(err)
        res.Status = "error"
        res.Error = err.Error()
    } else {
        Log.Debug(user)
        res.Status = "OK"
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

