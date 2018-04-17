package login

import (
    "context"

    "time"
    "github.com/sirupsen/logrus"
    "net/http"
    "encoding/json"
    "github.com/mongodb/mongo-go-driver/bson"
    "golang.org/x/crypto/bcrypt"
    "TwitterClone/user"
    "TwitterClone/wrappers"
)


type response struct {
    Status string `json:"status"`
    Error string `json:"error,omitempty"`
}

type userDetails struct {
    Username *string `json:"username"`
    Password *string `json:"password"`
}
var Log *logrus.Logger
func main() {
    Log.SetLevel(logrus.InfoLevel)
}


func authUser(details userDetails) bool {
    client, err := wrappers.NewClient()
    if err != nil {
        Log.Error("Mongodb error")
        return false
    }
    var user user.User
    dbStart := time.Now()
    db := client.Database("twitter")
    coll := db.Collection("users")
    filter := bson.NewDocument(bson.EC.String("username", *details.Username),
            bson.EC.Boolean("verified", true))
    err = coll.FindOne(
        context.Background(),
        filter).Decode(&user)
      elapsed := time.Since(dbStart)
      Log.WithFields(logrus.Fields{"endpoint" : "login","msg":"Check if user exists in DB time elapsed", "timeElapsed":elapsed.String()}).Info()
    if err != nil {
        return false
    }
    authed := bcrypt.CompareHashAndPassword([]byte(user.Password),
    []byte(*details.Password)) == nil
    return authed
}

func decodeRequest(r *http.Request) (userDetails, error) {
    decoder := json.NewDecoder(r.Body)
    var details userDetails
    err := decoder.Decode(&details)
    return details, err
}
func encodeResponse(w http.ResponseWriter, response interface{}) error {
    return json.NewEncoder(w).Encode(response)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
    timeStart := time.Now()
    var res response
    details, err := decodeRequest(r)
    if (err != nil) {
        res.Status = "error"
        res.Error = "Error decoding json."
    } else if validateDetails(details) {
        res = loginEndpoint(details)
    } else {
        res.Status = "error"
        res.Error = "Invalid request."
    }
    if (res.Status == "OK") {
        var cookie http.Cookie
        cookie.Name = "username"
        cookie.Value = *details.Username
        http.SetCookie(w, &cookie)
    }
    elapsed := time.Since(timeStart)
    Log.Info("elapsed: " + elapsed.String())
    encodeResponse(w, res)
}

func loginEndpoint(details userDetails) (response) {
    var res response
    // Check username and password against database.
    shouldLogin := authUser(details)
    if (shouldLogin) {
        res.Status = "OK"
    } else {
        res.Status = "error"
        res.Error = "User not found or incorrect password."
    }
    return res
}

func validateDetails(details userDetails) bool {
    return (details.Username != nil && details.Password != nil)
}
