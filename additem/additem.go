package main

import (
    "context"
    "net/http"
    "time"
    "github.com/onrik/logrus/filename"
    "os"
    "github.com/sirupsen/logrus"
    "encoding/json"
    "github.com/gorilla/mux"
    "github.com/mongodb/mongo-go-driver/mongo"
    "github.com/mongodb/mongo-go-driver/bson"
    "github.com/mongodb/mongo-go-driver/bson/objectid"
)

type Item struct {
    ID objectid.ObjectID
    Username string `json:"username"`
    Content *string `json:"content"`
    ChildType *string `json:"childType,omitempty"`
    Likes int32 `json:"likes"`
    Retweeted int32 `json:"retweeted"`
    Timestamp int64 `json:"timestamp"`
}

type response struct {
    Status string `json:"status"`
    ID string  `json:"id,omitempty"`
    Error string `json:"error,omitempty"`
}
var log = logrus.New()
func main() {
    r := mux.NewRouter()
    r.HandleFunc("/additem", addItemHandler).Methods("POST")
    http.Handle("/", r)
    log.AddHook(filename.NewHook())
    //log.SetLevel(log.InfoLevel)

    log.Out = os.Stdout
    //You could set this to any `io.Writer` such as a file
    file, err := os.OpenFile("logrus.log", os.O_CREATE|os.O_WRONLY, 0666)
    if err == nil {
     log.Out = file
    } else {
     log.Info("Failed to log to file, using default stderr")
    }
    log.Fatal(http.ListenAndServe(":8000", nil))
}


func checkLogin(r *http.Request) (string, error) {
    cookie, err := r.Cookie("username")
    if err != nil {
        return "", err
    } else {
        return cookie.Value, nil
    }
}


func insertItem(it *Item) (error, *objectid.ObjectID) {
    client, err := mongo.NewClient("mongodb://mongo.db:27017")
    if err != nil {
        log.Error("Error Connecting to Database")
        return err,nil
    }
    db := client.Database("twitter")
    col := db.Collection("tweets")
    id := objectid.New()
    log.Debug(id.Hex())
    it.ID = id
    it.Timestamp = time.Now().Unix()
    it.Likes = 0
    it.Retweeted = 0
    log.Debug(*it)
    doc := bson.NewDocument(bson.EC.ObjectID("_id", id),
            bson.EC.String("username", it.Username),
            bson.EC.String("content", *(it.Content)),
        bson.EC.Int32("likes", it.Likes),
        bson.EC.Int32("retweeted", it.Retweeted),
        bson.EC.Int64("timestamp", it.Timestamp))
    if it.ChildType != nil {
        doc.Append(bson.EC.String("childType", *(it.ChildType)))
    }
    _, err = col.InsertOne(
        context.Background(),
        doc)
    if err != nil {
      log.Error(err.Error())
        return err, nil
    } else {
        return nil, &id
    }
}

func decodeRequest(r *http.Request) (Item, error) {
    decoder := json.NewDecoder(r.Body)
    var it Item
    err := decoder.Decode(&it)
    return it, err
}

func encodeResponse(w http.ResponseWriter, response interface{}) error {
    return json.NewEncoder(w).Encode(response)
}

func addItemHandler(w http.ResponseWriter, r *http.Request) {
    var res response
    username, err := checkLogin(r)
    if err != nil {
      log.Error("User not logged in")
        res.Status = "error"
        res.Error = "User not logged in."
    } else {
        it, err := decodeRequest(r)
        if (err != nil) {
          log.Error("JSON decoding error")
            res.Status = "error"
            res.Error = "JSON decoding error."
        } else {
            it.Username = username
            res = addItemEndpoint(it)
        }
    }
    encodeResponse(w, res)
}

func addItemEndpoint(it Item) response {
    var res response
    log.Debug(it)
    valid := validateItem(it)
    if valid {
        // Add the Item.
        err, id := insertItem(&it)
        if err != nil {
            log.Error("Item could not be inserted into database.")
            res.Status = "error"
            res.Error = err.Error()
        } else {
            res.Status = "OK"
            res.ID = id.Hex()
        }
    } else {
        res.Status = "error"
        res.Error = "Invalid request."
        log.Info("Invalid request!")
    }
    log.WithFields(logrus.Fields{
    "content": it.Content,
    }).Info("About to return from adding the tweet")
    return res
}

func validateItem(it Item) bool {
    valid := true
    if (it.Content == nil) {
        valid = false
    } else if (it.ChildType == nil) {
        valid = true
    } else if (*it.ChildType != "retweet" && *it.ChildType != "reply") {
        // Invalid r
        log.Debug("childType not valid")
        valid = false
    }
    return valid
}
