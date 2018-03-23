package main

import (
    "context"
    "log"
    "net/http"
    "time"
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
    Retweeted int32 `json:retweeted"`
    Timestamp int64 `json:timestamp"`
}

type response struct {
    Status string `json:"status"`
    ID string  `json:"id,omitempty"`
    Error string `json:"error,omitempty"`
}

func main() {
    r := mux.NewRouter()
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    r.HandleFunc("/additem", addItemHandler).Methods("POST")
    http.Handle("/", r)
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


func insertItem(it *Item) *objectid.ObjectID {
    client, err := mongo.NewClient("mongodb://localhost:27017")
    if err != nil {
        log.Println("Error inserting")
        return nil
    }
    db := client.Database("twitter")
    col := db.Collection("tweets")
    id := objectid.New()
    log.Println(id)
    it.ID = id
    it.Timestamp = time.Now().Unix()
    it.Likes = 0
    it.Retweeted = 0
    log.Println(*it)
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
        return nil
    } else {
        return &id
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
        res.Status = "error"
        res.Error = "User not logged in."
    } else {
        it, err := decodeRequest(r)
        if (err != nil) {
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
    log.Println(it)
    valid := validateItem(it)
    if valid {
        // Add the Item.
        id := insertItem(&it)
        if id == nil {
            res.Status = "error"
            res.Error = "Item could not be inserted into database."
        } else {
            res.Status = "OK"
            res.Error = ""
            res.ID = id.Hex()
        }
        log.Println("Encoded!")
    } else {
        res.Status = "error"
        res.Error = "Invalid request."
        log.Println("Not valid!")
    }
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
        log.Println("childType not valid")
        valid = false
    }
    return valid
}
