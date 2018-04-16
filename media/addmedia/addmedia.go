package addmedia

import (
    "context"
    "net/http"
    "time"
    
    "io"
    "bytes"
    "github.com/sirupsen/logrus"
    "encoding/json"
    "github.com/mongodb/mongo-go-driver/bson/objectid"
    "TwitterClone/wrappers"
    "TwitterClone/media"
)

type response struct {
    Status string `json:"status"`
    ID string  `json:"id,omitempty"`
    Error string `json:"error,omitempty"`
}

var Log *logrus.Logger
func main() {
    Log.SetLevel(logrus.ErrorLevel)
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

func AddMediaHandler(w http.ResponseWriter, r *http.Request) {
    var res response
    username, err := checkLogin(r)
    if err != nil {
        Log.Error(err)
        encodeResponse(w, errResponse(err))
        return
    }
    content, header, err := r.FormFile("content") // Get binary payload.
    if err != nil {
        Log.Error(err)
        encodeResponse(w, errResponse(err))
        return
    }
    defer content.Close()
    Log.Debug(header.Header)
    bufContent := bytes.NewBuffer(nil)
    if _, err := io.Copy(bufContent, content); err != nil {
        Log.Error(err)
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
        Log.Error(err)
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
    Log.Info("Time elapsed: " + elapsed.String())
    if err != nil {
        Log.Error(err.Error())
        return nilObjectID, err
    } else {
        return id, nil
    }
}
