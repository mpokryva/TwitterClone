package media_endpoints

import (
    "context"
    "net/http"
    "time"
    "errors"
    "strconv"
    "github.com/sirupsen/logrus"
    "github.com/gorilla/mux"
    "github.com/mongodb/mongo-go-driver/bson/objectid"
    "github.com/mongodb/mongo-go-driver/bson"
    "TwitterClone/wrappers"
    "TwitterClone/media"
)

var Log *logrus.Logger
func main() {
    Log.SetLevel(logrus.ErrorLevel)
}



func GetMediaHandler(w http.ResponseWriter, r *http.Request) {
  start := time.Now()
    vars := mux.Vars(r)
    id := vars["id"]
    Log.Debug(id)
    var media media.Media
    oid, err := objectid.FromHex(id)
    if err != nil {
        Log.Error(err)
    } else {
        media, err = getMedia(oid)
        if err != nil {
            Log.Error(err)
        }
    }

    elapsed := time.Since(start)
    Log.Info("AddItem elapsed: " + elapsed.String())
    encodeResponse(w, media)
}

func getMedia(oid objectid.ObjectID) (media.Media, error) {
  dbStart := time.Now()
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
    elapsed := time.Since(dbStart)
    Log.WithFields(logrus.Fields{"timeElapsed":elapsed.String()}).Info("Get Media time elapsed")
    return media, err
}

func encodeResponse(w http.ResponseWriter, m media.Media) {
    if (m.Content == nil) {
        err := errors.New("Media not found.")
        Log.Debug(err)
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    w.Header().Set("Content-Type", m.Header.Header["Content-Type"][0])
    w.Header().Set("Content-Length", strconv.Itoa(len(m.Content)))
    w.Write(m.Content)
}
