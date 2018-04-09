package main

import (
    "context"
    "net/http"
    "time"
    "os"
    "errors"
    "github.com/sirupsen/logrus"
    "encoding/json"
    "github.com/gorilla/mux"
    "github.com/mongodb/mongo-go-driver/bson/objectid"
    "TwitterClone/wrappers"
    "TwitterClone/media"
)

type request struct {
    Content []byte `json:"content"`
}

type response struct {
    Status string `json:"status"`
    ID string  `json:"id,omitempty"`
    Error string `json:"error,omitempty"`
}

var log *logrus.Logger
func main() {
    r := mux.NewRouter()
    r.HandleFunc("/addmedia", addItemHandler).Methods("POST")
    http.Handle("/", r)
    // Log to a file
    var f *os.File
    var err error
    log, f, err = wrappers.FileLogger("addmedia.log", os.O_CREATE | os.O_RDWR,
        0666)
    if err != nil {
        log.Fatal("Logging file could not be opened.")
    }
    f.Truncate(0)
    f.Seek(0, 0)
    defer f.Close()
    log.SetLevel(logrus.DebugLevel)
    log.Fatal(http.ListenAndServe(":8011", nil))
}


func checkLogin(r *http.Request) (string, error) {
    cookie, err := r.Cookie("username")
    if err != nil {
        return "", err
    } else {
        return cookie.Value, nil
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
    _, header, err := r.FormFile("content")
    log.Debug(header)
    req, err := decodeRequest(r)
    if (err != nil) {
        log.Error("JSON decoding error")
        res.Status = "error"
        res.Error = "JSON decoding error."
        return
    }
    err = validateRequest(req)
    if err == nil {
        var m media.Media
        m.Content = req.Content
        m.Username = username
        res = addMediaEndpoint(m)
    } else {
        res.Status = "error"
        res.Error = err.Error()
    }
    encodeResponse(w, res)
}

func addMediaEndpoint(m media.Media) response {
    var res response
    log.Debug(m)
    // Add the Media.
    oid, err := insertMedia(m)
    if err != nil {
        log.Error(err)
        res.Status = "error"
        res.Error = err.Error()
    } else {
        res.Status = "OK"
        res.ID = oid.Hex()
    }
    return res
}

func insertMedia(m media.Media) (objectid.ObjectID, error) {
    start := time.Now()
    client, err := wrappers.NewClient()
    if err != nil {

    }
    db := client.Database("twitter")
    col := db.Collection("media")
    id := objectid.New()
    m.ID = id
    log.Debug(m)
    _, err = col.InsertOne(context.Background(), &m)
    elapsed := time.Since(start)
    log.Info("Time elapsed: " + elapsed.String())
    var nilObjectID objectid.ObjectID
    if err != nil {
        log.Error(err.Error())
        return nilObjectID, err
    } else {
        return id, nil
    }
}
func validateRequest(req request) error {
    if req.Content != nil {
        return nil
    } else {
        return errors.New("Media content is null.")
    }
}
