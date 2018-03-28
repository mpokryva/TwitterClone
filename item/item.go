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

type Item struct {
    ID string `json:"_id"`
    Content string `json:"content"`
    Username string `json:"username"`
    Property property `json:"property"`
    Retweeted int `json:"retweeted"`
    Timestamp int64 `json:"timestamp"`
}

type property struct {
  Likes int `json:"likes"`
}

type response struct {
    Status string `json:"status"`
    Item Item `json:"item,omitempty"`
    Error string `json:"error,omitempty"`
}

func main() {
    r := mux.NewRouter()
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    r.HandleFunc("/item/{id}", getItemHandler).Methods("GET")
    // r.HandleFunc("/item", getItemHandler).Methods("GET")

    http.Handle("/", r)
    log.Fatal(http.ListenAndServe(":8005", nil))
}

func getItemHandler(w http.ResponseWriter, r *http.Request) {
    var res response
    id := mux.Vars(r)["id"]
    log.Println(id)
    res = getItemEndpoint(id)
    encodeResponse(w, res)
}

func getItemEndpoint(id string) response {    
    return getItem(id)
}

func getItem(id string) response {
    var info Item
    var resp response 
    var prop property
    client, err := mongo.NewClient("mongodb://localhost:27017")
    if err != nil {
        log.Println("Error connecting to database")
        resp.Status = "error"
        resp.Error = "Database unavailable"
        return resp
    }
    db := client.Database("twitter")
    col := db.Collection("tweets")
    objectid,err := objectid.FromHex(id)
    if err != nil {
        // log.Println(err)
        // log.Println(objectid)
        resp.Status = "error"
        resp.Error = "Invalid Item ID"
        return resp
    }

    doc := bson.NewDocument(bson.EC.ObjectID("_id", objectid))
    item, err := col.Find(
        context.Background(),
        doc)
    if err != nil {
        log.Println("Did not find ObjectID")
        resp.Status = "error"
        resp.Error = "Invalid Item ID"
        return resp
    }
    // log.Println(item)
    for item.Next(context.Background()){
        row := bson.NewDocument()
        err = item.Decode(row)
        // log.Println(info)
        // log.Println(row)
 
        res, err4 := row.Lookup("content")
        if err4 == nil{
          info.Content = res.Value().StringValue()
        }
        res, err4 = row.Lookup("_id")
        if err4 == nil{
          info.ID = res.Value().ObjectID().Hex()
        }
        res, err4 = row.Lookup("username")
        if err4 == nil{
          info.Username = res.Value().StringValue()
        }
        res, err4 = row.Lookup("likes")
        if err4 == nil{
          prop.Likes = (int)(res.Value().Int32())
          info.Property = prop
        }
        res, err4 = row.Lookup("retweeted")
        if err4 == nil{
          info.Retweeted = (int)(res.Value().Int32())
        }
        res, err4 = row.Lookup("timestamp")
        if err4 == nil{
          info.Timestamp= res.Value().Int64()
        }
    }
    if info.Username == "" {
      resp.Status = "error"
      resp.Error = "Item could not be found."
    } else {
      resp.Status = "OK"
      resp.Item = info
      resp.Error = ""
    }
      log.Println("Encoded!")
      return resp
}

func encodeResponse(w http.ResponseWriter, response interface{}) error {
    return json.NewEncoder(w).Encode(response)
}

func validateItem(it Item) bool {
    valid := true
    return valid
}
