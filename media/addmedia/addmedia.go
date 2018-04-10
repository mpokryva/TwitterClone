package main

import (
    "context"
    "net/http"
    "time"
    "os"
    "io"
    "bytes"
    "github.com/sirupsen/logrus"
    "encoding/json"
    "github.com/gorilla/mux"
    "github.com/mongodb/mongo-go-driver/bson/objectid"
    "TwitterClone/wrappers"
    "TwitterClone/media"
)

type response struct {
    Status string `json:"status"`
    ID string  `json:"id,omitempty"`
    Error string `json:"error,omitempty"`
}

var log *logrus.Logger
func main() {
    r := mux.NewRouter()
    r.HandleFunc("/addmedia", addMediaHandler).Methods("POST")
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

func encodeResponse(w http.ResponseWriter, response interface{}) error {
    return json.NewEncoder(w).Encode(response)
}

func errResponse(err error) response {
    var res response
    res.Status = "error"
    res.Error = err.Error()
    return res
}

func addMediaHandler(w http.ResponseWriter, r *http.Request) {
    var res response
    username, err := checkLogin(r)
    if err != nil {
        log.Error(err)
        encodeResponse(w, errResponse(err))
        return
    }
    content, header, err := r.FormFile("content") // Get binary payload.
    if err != nil {
        log.Error(err)
        encodeResponse(w, errResponse(err))
        return
    }
    defer content.Close()
    log.Debug(header.Header)
    bufContent := bytes.NewBuffer(nil)
    if _, err := io.Copy(bufContent, content); err != nil {
        log.Error(err)
        encodeResponse(w, errResponse(err))
        return
    }
    buf := bufContent.Bytes()
    var m media.Media
    if header != nil {
        m.Header = *header
    }
    m.Content = buf
    m.Username = username
    res = addMediaEndpoint(m)
    encodeResponse(w, res)
}

func addMediaEndpoint(m media.Media) response {
    var res response
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
    var nilObjectID objectid.ObjectID
    start := time.Now()
    client, err := wrappers.NewClient()
    if err != nil {
        return nilObjectID, err
    }
    db := client.Database("twitter")
    col := db.Collection("media")
    id := objectid.New()
    m.ID = id
    _, err = col.InsertOne(context.Background(), &m)
    elapsed := time.Since(start)
    log.Info("Time elapsed: " + elapsed.String())
    if err != nil {
        log.Error(err.Error())
        return nilObjectID, err
    } else {
        return id, nil
    }
}
