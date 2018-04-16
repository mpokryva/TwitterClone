package follow

import (
    "context"
    "net/http"
    "os"
    "errors"
    "github.com/sirupsen/logrus"
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
var log *logrus.Logger
func main() {
    r := mux.NewRouter()
    r.HandleFunc("/follow", FollowHandler).Methods("POST")
    http.Handle("/", r)
    // Log to a file
    var f *os.File
    var err error
    log, f, err = wrappers.FileLogger("follow.log", os.O_CREATE | os.O_RDWR,
        0666)
    if err != nil {
        Log.Fatal("Logging file could not be opened.")
    }
    f.Truncate(0)
    f.Seek(0, 0)
    defer f.Close()
    Log.SetLevel(logrus.ErrorLevel)
    Log.Fatal(http.ListenAndServe(":8009", nil))
}

func checkLogin(r *http.Request) (string, error) {
    cookie, err := r.Cookie("username")
    if err != nil {
        return "", err
    } else {
        return cookie.Value, nil
    }
}

func followUser(currentUser string, userToFol string, follow bool) error {
    client, err := wrappers.NewClient()
    if err != nil {
        return nil
    }

    db := client.Database("twitter")
    coll := db.Collection("users")
    // Check if user to follow exists.
    // Assuming that logged in user exists (not bogus cookie).
    checkUserFilter := bson.NewDocument(
        bson.EC.String("username", userToFol))
    var userToFollow user.User
    err = coll.FindOne(context.Background(), checkUserFilter).Decode(&userToFollow)
    if err != nil {
        Log.Info(err)
        return errors.New("User to follow doesn't exist.")
    }
    var listOp string
    var countInc int32
    if follow {
        listOp = "$addToSet"
        countInc = 1
    } else {
        listOp = "$pull"
        countInc = -1
    }
    // Update following list of logged in user.
    filter := bson.NewDocument(
        bson.EC.String("username", currentUser))
    update := bson.NewDocument(
        bson.EC.SubDocumentFromElements(listOp,
            bson.EC.String("following", userToFol)))
    err = UpdateOne(coll, filter, update)
    if err != nil {
        return err
    }

    // Update following count.
    update = bson.NewDocument(
        bson.EC.SubDocumentFromElements("$inc",
            bson.EC.Int32("followingCount", countInc)))
    err = UpdateOne(coll, filter, update)
    if err != nil {
        return err
    }

    // Updated following successfully. Now updating followers of other user.
    filter = bson.NewDocument(
        bson.EC.String("username", userToFol))
    update = bson.NewDocument(
        bson.EC.SubDocumentFromElements(listOp,
            bson.EC.String("followers", currentUser)))
    err = UpdateOne(coll, filter, update)
    if err != nil {
        return err
    }
    // Update follower count.
    update = bson.NewDocument(
        bson.EC.SubDocumentFromElements("$inc",
            bson.EC.Int32("followerCount", countInc)))
    return UpdateOne(coll, filter, update)
}

func UpdateOne(coll *mongo.Collection, filter interface{}, update interface{}) error {
    result, err := coll.UpdateMany( // UpdateMany is temporary.
        context.Background(),
        filter, update)
    var success = false
    if result != nil {
        Log.Debug(*result)
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

func FollowHandler(w http.ResponseWriter, r *http.Request) {
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
            Log.WithFields(logrus.Fields{
                "username": *it.Username,
                "follow": *it.Follow,
                "currentUser": username}).Info()
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
        err := followUser(username, *it.Username, *it.Follow)
        if err != nil {
            res.Status = "error"
            res.Error = err.Error()
        } else {
            res.Status = "OK"
        }
    } else {
        res.Status = "error"
        res.Error = "Invalid request."
        Log.Info("Invalid request!")
    }
    return res
}

func validateReq(it Request) bool {
    return it.Username != nil && it.Follow != nil
}
