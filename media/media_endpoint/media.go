package main

import (
    "context"
    "net/http"
    "os"
    "errors"
    "strconv"
    "github.com/sirupsen/logrus"
    "github.com/gorilla/mux"
    "github.com/mongodb/mongo-go-driver/bson/objectid"
    "github.com/mongodb/mongo-go-driver/bson"
    "TwitterClone/wrappers"
    "TwitterClone/media"
)

var log *logrus.Logger
func main() {
    r := mux.NewRouter()
    r.HandleFunc("/media/{id}", mediaHandler).Methods("GET")
    http.Handle("/", r)
    // Log to a file
    var f *os.File
    var err error
    log, f, err = wrappers.FileLogger("media.log", os.O_CREATE | os.O_RDWR,
        0666)
    if err != nil {
        log.Fatal("Logging file could not be opened.")
    }
    f.Truncate(0)
    f.Seek(0, 0)
    defer f.Close()
    log.SetLevel(logrus.ErrorLevel)
    log.Fatal(http.ListenAndServe(":8010", nil))
}

func mediaHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    log.Debug(id)
    var media media.Media
    oid, err := objectid.FromHex(id)
    if err != nil {
        log.Error(err)
    } else {
        media, err = getMedia(oid)
        if err != nil {
            log.Error(err)
        }
    }
    encodeResponse(w, media)
}

func getMedia(oid objectid.ObjectID) (media.Media, error) {
    var nilMedia media.Media
    client, err := wrappers.NewClient()
    if err != nil {
        return nilMedia, err
    }
    db := client.Database("twitter")
    coll := db.Collection("media")
    var media media.Media
    filter := bson.NewDocument(bson.EC.ObjectID("_id", oid))
    err = coll.FindOne(context.Background(), filter).Decode(&media)
    return media, err
}

func encodeResponse(w http.ResponseWriter, m media.Media) {
    if (m.Content == nil) {
        err := errors.New("Media not found.")
        log.Debug(err)
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    w.Header().Set("Content-Type", m.Header.Header["Content-Type"][0])
    w.Header().Set("Content-Length", strconv.Itoa(len(m.Content)))
    w.Write(m.Content)
}
