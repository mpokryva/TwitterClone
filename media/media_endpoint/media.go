package main

import (
    "context"
    "net/http"
    "os"
    "github.com/sirupsen/logrus"
    "encoding/json"
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
    log.SetLevel(logrus.DebugLevel)
    log.Fatal(http.ListenAndServe(":8010", nil))
}

func mediaHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    log.Debug(id)
    var media []byte
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

func getMedia(oid objectid.ObjectID) ([]byte, error) {
    client, err := wrappers.NewClient()
    if err != nil {
        return nil, err
    }
    db := client.Database("twitter")
    coll := db.Collection("media")
    var media media.Media
    filter := bson.NewDocument(bson.EC.ObjectID("_id", oid))
    err = coll.FindOne(context.Background(), filter).Decode(&media)
    log.Debug(media)
    return media.Content, err
}

func encodeResponse(w http.ResponseWriter, response interface{}) error {
    w.Header().Set("Content-Type", "application/octet-stream")
    return json.NewEncoder(w).Encode(response)
}


