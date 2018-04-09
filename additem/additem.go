package main

import (
    "context"
    "net/http"
    "time"
    "os"
    "reflect"
    "github.com/sirupsen/logrus"
    "encoding/json"
    "github.com/gorilla/mux"
    "github.com/mongodb/mongo-go-driver/bson/objectid"
    "TwitterClone/wrappers"
    "TwitterClone/item"
)

type request struct {
    Content *string `json:"content"`
    ChildType *string `json:"childType,omitempty"`
}

type response struct {
    Status string `json:"status"`
    ID string  `json:"id,omitempty"`
    Error string `json:"error,omitempty"`
}

var log *logrus.Logger
func main() {
    r := mux.NewRouter()
    r.HandleFunc("/additem", addItemHandler).Methods("POST")
    http.Handle("/", r)
    // Log to a file
    var f *os.File
    var err error
    log, f, err = wrappers.FileLogger("additem.log", os.O_CREATE | os.O_RDWR,
        0666)
    if err != nil {
        log.Fatal("Logging file could not be opened.")
    }
    f.Truncate(0)
    f.Seek(0, 0)
    defer f.Close()
    log.SetLevel(logrus.DebugLevel)
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


func insertItem(it item.Item) (error, *objectid.ObjectID) {
    start := time.Now()
    client, err := wrappers.NewClient()
    db := client.Database("twitter")
    col := db.Collection("tweets")
    id := objectid.New()
    log.Debug(id.Hex())
    it.ID = id
    log.Info(reflect.TypeOf(objectid.ObjectID{}))
    it.Timestamp = time.Now().Unix()
    it.Property.Likes = 0
    it.Retweeted = 0
    log.Debug(it)
    _, err = col.InsertOne(context.Background(), &it)
    elapsed := time.Since(start)
    log.Info("Time elapsed: " + elapsed.String())
    if err != nil {
      log.Error(err.Error())
        return err, nil
    } else {
        return nil, &id
    }
}

func decodeRequest(r *http.Request) (request, error) {
    decoder := json.NewDecoder(r.Body)
    var req request
    err := decoder.Decode(&req)
    return req, err
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
        return
    }
    req, err := decodeRequest(r)
    if (err != nil) {
        log.Error("JSON decoding error")
        res.Status = "error"
        res.Error = "JSON decoding error."
        return
    }
    valid := validateRequest(req)
    if valid {
        var it item.Item
        it.Content = *req.Content
        if req.ChildType != nil {
            it.ChildType = *req.ChildType
        }
        it.Username = username
        res = addItemEndpoint(it)
    } else {
        res.Status = "error"
        res.Error = "Invalid request."
    }
    encodeResponse(w, res)
}

func addItemEndpoint(it item.Item) response {
    var res response
    log.Debug(it)
    // Add the Item.
    err, id := insertItem(it)
    if err != nil {
        log.Error("Item could not be inserted into database.")
        res.Status = "error"
        res.Error = err.Error()
    } else {
        res.Status = "OK"
        res.ID = id.Hex()
    }
    return res
}

func validateRequest(req request) bool {
    valid := true
    if (req.Content == nil) {
        valid = false
    } else if (req.ChildType == nil) {
        valid = true
    } else if (*req.ChildType != "retweet" && *req.ChildType != "reply") {
        // Invalid r
        log.Debug("childType not valid")
        valid = false
    }
    return valid
}
