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
    "github.com/mongodb/mongo-go-driver/bson"
    "github.com/mongodb/mongo-go-driver/mongo"
    "TwitterClone/wrappers"
    "TwitterClone/item"
)

type request struct {
    Content *string `json:"content"`
    ChildType *string `json:"childType,omitempty"`
    ParentID *string `json:"parent,omitempty"`
    MediaIDs *[]string `json:"media,omitempty"`
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


func insertItem(it item.Item) (objectid.ObjectID, error) {
    start := time.Now()
    client, err := wrappers.NewClient()
    db := client.Database("twitter")
    col := db.Collection("tweets")
    oid := objectid.New()
    log.Debug(oid.Hex())
    it.ID = oid
    it.Timestamp = time.Now().Unix()
    log.Debug(it)
    var nilObjectID objectid.ObjectID
    _, err = col.InsertOne(context.Background(), &it)
    elapsed := time.Since(start)
    log.Info("Time elapsed: " + elapsed.String())
    if err != nil {
         log.Error(err)
        return nilObjectID, err
    }
    // Update media which item references.
    var result *mongo.UpdateResult
    if it.MediaIDs != nil {
        col = db.Collection("media")
        bArray := bson.NewArray()
        for _, mOID := range it.MediaIDs {
            bArray.Append(bson.VC.ObjectID(mOID))
        }
        filter := bson.NewDocument(
            bson.EC.SubDocumentFromElements("_id",
            bson.EC.Array("$in", bArray)))
        update := bson.NewDocument(
            bson.EC.SubDocumentFromElements("$addToSet",
            bson.EC.ObjectID("item_ids", oid)))
            result, err = col.UpdateMany(context.Background(), filter, update)
            log.Debug(result)
    }
    if err != nil {
        log.Error(err)
        return nilObjectID, nil
    } else {
        return oid, nil
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
    pOID, mOIDs, err := validateRequest(req)
    if err == nil {
        var it item.Item
        it.Content = *req.Content
        if req.ChildType != nil {
            it.ChildType = *req.ChildType
        }
        if req.ParentID != nil {
            it.ParentID = pOID
            log.Debug(*req.ParentID)
        }
        if req.MediaIDs != nil {
            it.MediaIDs = mOIDs
        }
        it.Username = username
        res = addItemEndpoint(it)
    } else {
        res.Status = "error"
        res.Error = err.Error()
    }
    encodeResponse(w, res)
}

func addItemEndpoint(it item.Item) response {
    var res response
    log.Debug(it)
    // Add the Item.
    id, err := insertItem(it)
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

func validateRequest(req request) (objectid.ObjectID, []objectid.ObjectID, error) {
    var pOID objectid.ObjectID
    var mOIDs []objectid.ObjectID
    var err error
    if req.Content == nil {
        err = errors.New("Null content")
    } else if req.ChildType == nil {
        if req.ParentID != nil {
            err = errors.New("Parent not null when childType is")
        }
    } else if (*req.ChildType != "retweet" && *req.ChildType != "reply") {
        err = errors.New("Child type not valid")
    } else if req.ParentID == nil {
        err = errors.New("Parent must be set when child type exists.")
    } else {

        pOID, err = objectid.FromHex(*req.ParentID)
    }
    if err == nil && req.MediaIDs != nil {
        for _, mID := range *req.MediaIDs {
            mOID, idErr := objectid.FromHex(mID)
            if err == nil {
                mOIDs = append(mOIDs, mOID)
            } else {
                err = idErr
                break
            }
        }
    }
    return pOID, mOIDs, err
}
