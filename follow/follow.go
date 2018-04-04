package main

import (
    "context"
    "net/http"
    "errors"
    "github.com/onrik/logrus/filename"
    log "github.com/sirupsen/logrus"
    "encoding/json"
    "github.com/gorilla/mux"
    "github.com/mongodb/mongo-go-driver/bson"
    "github.com/mongodb/mongo-go-driver/mongo"
    "TwitterClone/wrappers"
    "TwitterClone/user"
)

type Request struct {
    Username *string `json:"username"`
    Follow *bool `json:"follow"`
}

type response struct {
    Status string `json:"status"`
    Error string `json:"error,omitempty"`
}

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/follow", followHandler).Methods("POST")
    http.Handle("/", r)
    log.AddHook(filename.NewHook())
    log.SetLevel(log.DebugLevel)
    log.Fatal(http.ListenAndServe(":8009", nil))
}


func checkLogin(r *http.Request) (string, error) {
    cookie, err := r.Cookie("username")
    if err != nil {
        return "", err
    } else {
        return cookie.Value, nil
    }
}

func followUser(username string, it *Request) error {
    client, err := wrappers.NewClient()
    if err != nil {
        return nil
    }

    db := client.Database("twitter")
    coll := db.Collection("users")
    // Check if user to follow exists.
    // Assuming that logged in user exists (not bogus cookie).
    checkUserFilter := bson.NewDocument(
        bson.EC.String("username", *it.Username))
    var userToFollow user.User
    err = coll.FindOne(context.Background(), checkUserFilter).Decode(&userToFollow)
    if err != nil {
        log.Info(err)
        return errors.New("User to follow doesn't exist.")
    }
    // Update following list of logged in user.
    filter := bson.NewDocument(
        bson.EC.String("username", username))
    //arr := bson.EC.ArrayFromElements(username,
            //bson.VC.String(*it.Username))
    update := bson.NewDocument(
        bson.EC.SubDocumentFromElements("$addToSet",
            bson.EC.String("following", *it.Username)))
    err = UpdateOne(coll, filter, update)
    if err != nil {
        return err
    }
    // Updated following successfully. Now updating followers of other user.
    filter = bson.NewDocument(
        bson.EC.String("username", *it.Username))
    update = bson.NewDocument(
        bson.EC.SubDocumentFromElements("$addToSet",
            bson.EC.String("followers", username)))
    return UpdateOne(coll, filter, update)
}

func UpdateOne(coll *mongo.Collection, filter interface{}, update interface{}) error {
    result, err := coll.UpdateOne(
        context.Background(),
        filter, update)
    var success = false
    if result != nil {
        log.Debug(*result)
        success = result.ModifiedCount == 1
    }
    if err != nil {
        return err
    } else if !success {
        return errors.New("Database is operating normally, but follow update " +
        "operation failed.")
    } else {
        return nil;
    }
}

func decodeRequest(r *http.Request) (Request, error) {
    decoder := json.NewDecoder(r.Body)
    var it Request
    err := decoder.Decode(&it)
    return it, err
}

func encodeResponse(w http.ResponseWriter, response interface{}) error {
    return json.NewEncoder(w).Encode(response)
}

func followHandler(w http.ResponseWriter, r *http.Request) {
    var res response
    username, err := checkLogin(r)
    if err != nil {
        res.Status = "error"
        res.Error = "User not logged in."
    } else {
        it, err := decodeRequest(r)
        if (err != nil) {
            res.Status = "error"
            res.Error = "JSON decoding error."
        } else {
            log.WithFields(log.Fields{
                "username": *it.Username,
                "follow": *it.Follow}).Info()
            res = followEndpoint(username, it)
        }
    }
    encodeResponse(w, res)
}

func followEndpoint(username string,it Request) response {
    var res response
    valid := validateReq(it)
    if valid {
        // Add the Item.
        err := followUser(username,&it)
        if err != nil {
            res.Status = "error"
            res.Error = err.Error()
        } else {
            res.Status = "OK"
        }
    } else {
        res.Status = "error"
        res.Error = "Invalid request."
        log.Info("Invalid request!")
    }
    return res
}

func validateReq(it Request) bool {
    return it.Username != nil && it.Follow != nil
}
