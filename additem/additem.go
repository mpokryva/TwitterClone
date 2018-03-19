package main

import (
    "context"
    "log"
    "net/http"
    "encoding/json"
    "github.com/gorilla/mux"
    "github.com/mongodb/mongo-go-driver/mongo"
    "github.com/mongodb/mongo-go-driver/bson"
    "github.com/mongodb/mongo-go-driver/bson/objectid"
)

type item struct {
    ID objectid.ObjectID
    Content *string `json:"content"`
    ChildType *string `json:"childType"`
}

type response struct {
    Status string `json:"status"`
    ID string  `json:"id"`
    Error string `json:"error"`
}

func main() {
    r := mux.NewRouter()
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    r.HandleFunc("/additem", addItem).Methods("POST")
    http.Handle("/", r)
    http.ListenAndServe(":8080", nil)
}

func insertItem(it *item) *objectid.ObjectID {
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
    log.Println(*it)
    doc := bson.NewDocument(bson.EC.ObjectID("_id", id),
            bson.EC.String("content", *(it.Content)))
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

func addItem(w http.ResponseWriter, req *http.Request) {
    decoder := json.NewDecoder(req.Body)
    var it item
    err := decoder.Decode(&it)
    if err != nil {
        panic(err)
    }
    log.Println(it)
    valid := validateItem(it)
    if valid {
        // Add the item.
        log.Println(it)
        id := insertItem(&it)
        var res response
        if id == nil {
            res.Status = "error"
            res.Error = "Item could not be inserted into database."
        } else {
            res.Status = "OK"
            res.Error = ""
            log.Println(id)
            res.ID = id.Hex()
        }
        //w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(res)
        log.Println("Encoded!")
    } else {
        log.Println("Not valid!")
    }
}

func validateItem(it item) bool {
    valid := true
    if (it.Content == nil) {
        valid = false
    } else if (it.ChildType == nil) {
        valid = true
    } else if (*it.ChildType != "retweet" && *it.ChildType != "reply") {
        // Invalid req
        log.Println("childType not valid")
        valid = false
    }
    return valid
}
