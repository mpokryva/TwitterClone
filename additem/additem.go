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
    ChildType *string `json:"childType,omitempty"`
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


func isLoggedIn(r *http.Request) bool {
    cookie, _ := r.Cookie("username")
    return cookie != nil
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

func decodeRequest(r *http.Request) (item, error) {
    decoder := json.NewDecoder(r.Body)
    var it item
    err := decoder.Decode(&it)
    return it, err
}

func encodeResponse(w http.ResponseWriter, response interface{}) error {
    return json.NewEncoder(w).Encode(response)
}

func addItemHandler(w http.ResponseWriter, r *http.Request) {
    var res response
    if !isLoggedIn(r) {
        res.Status = "error"
        res.Error = "User not logged in."
    } else {
        item, err := decodeRequest(r)
        if (err != nil) {
            res.Status = "error"
            res.Error = "JSON decoding error."
        } else {
            res = addItemEndpoint(item)
        }
    }
    encodeResponse(w, res)
}

func addItemEndpoint(it item) response {
    var res response
    log.Println(it)
    valid := validateItem(it)
    if valid {
        // Add the item.
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

func validateItem(it item) bool {
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
