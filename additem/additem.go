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
    Content *string
    ChildType *string
}

func main() {
    r := mux.NewRouter()
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    r.HandleFunc("/additem", addItem).Methods("POST")
    http.Handle("/", r)
    http.ListenAndServe(":8080", nil)
}

func insertItem(it *item) {
    client, err := mongo.NewClient("mongodb://localhost:27017")
    if err != nil {
        log.Println("Panicking")
        panic(err)
    }
    db := client.Database("twitter")
    col := db.Collection("tweets")
    id := objectid.New()
    log.Println(id)
    it.ID = id
    log.Println(*it)
    col.InsertOne(
        context.Background(),
        bson.NewDocument(bson.EC.String("item", "test")))
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
        insertItem(&it)
    } else {

    }
    // Validate req
    /*for {
        var it item
        if err := decoder.Decode(&it); err == io.EOF {
            break
        } else if err != nil {
            panic(err)
        }
    }*/
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
