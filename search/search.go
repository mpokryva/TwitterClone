package main

import (
    "time"
    "context"
    "log"
    "net/http"
    //"net/http/httputil"
    "encoding/json"
    "github.com/gorilla/mux"
    "github.com/mongodb/mongo-go-driver/mongo"
    "github.com/mongodb/mongo-go-driver/bson"
    //"github.com/mongodb/mongo-go-driver/bson/objectid"
)

type params struct {
    Timestamp int64 `json:"timestamp,string"`
    Limit int `json:"limit,string"`
}

type Item struct {
    ID string`json:"id"`
    Content string `json:"content"`
    Username string `json:"username"`
    Property property `json:"property"`
    Retweeted int `json:"retweeted"`
    Timestamp int64 `json:"timestamp"`
}
type property struct {
  Likes int `json:"likes"`
}
type res struct {
  Status string `json:"status"`
  Items []Item `json:"items"`
  Error string `json:"error,omitempty"`
}

func main() {
    r := mux.NewRouter()
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    r.HandleFunc("/search", search).Methods("POST")
    http.Handle("/", r)
    http.ListenAndServe(":8006", nil)
}

func search(w http.ResponseWriter, req *http.Request) {
    //debug, err := httputil.DumpRequest(req, true)
    decoder := json.NewDecoder(req.Body)
    var start params
    var r res
    err := decoder.Decode(&start)
    if err != nil {
       r.Status = "error"
       r.Error = err.Error()
       json.NewEncoder(w).Encode(r)
       return
    }
    if(start.Timestamp == 0){
      start.Timestamp = time.Now().Unix()
    }
    if(start.Limit == 0){
      start.Limit = 25
    }
    if(start.Limit > 100){
      r.Status = "error"
      r.Error = "Limit must be under 100"
    }

    itemList, err := generateList(start)
    if (err == nil) {
      //it worked
      r.Status = "OK"
      r.Items = itemList
    } else {
        r.Status = "error"
        r.Error = err.Error()
    }

  json.NewEncoder(w).Encode(r)
}

func generateList(sPoint params) ([]Item, error){
  client, err := mongo.NewClient("mongodb://localhost:27017")
  if err != nil {
      log.Println("Error Connecting")
      return nil, err
  }
  db := client.Database("twitter")
  col := db.Collection("tweets")

  var tweetList []Item
  var info Item
  var prop property
  doc := bson.NewDocument(bson.EC.SubDocumentFromElements("timestamp",bson.EC.Int64("$lte", (int64)(sPoint.Timestamp),)))
  set,err := col.Find(
      context.Background(),
      doc)
  if err != nil {
    log.Println("Error querying")
      return nil, err
  } else {
    lim := sPoint.Limit
    for set.Next(context.Background()) && lim>0{

      row := bson.NewDocument()
      //err = set.Decode(info)
      err = set.Decode(row)
      //log.Println(row)

      res, err4 := row.Lookup("content")
      if err4 == nil{
        info.Content = res.Value().StringValue()
      }

      res, err4 = row.Lookup("_id")
      if err4 == nil{
        info.ID = res.Value().ObjectID().Hex()
      }

      res, err4 = row.Lookup("likes")
      if err4 == nil{
        prop.Likes = (int)(res.Value().Int32())
        info.Property = prop
      }

      res, err4 = row.Lookup("username")
      if err4 == nil{
        info.Username = res.Value().StringValue()
      }

      res, err4 = row.Lookup("retweeted")
      if err4 == nil{
        info.Retweeted = (int)(res.Value().Int32())
      }

      res, err4 = row.Lookup("timestamp")
      if err4 == nil{
        info.Timestamp= res.Value().Int64()
      }

      tweetList = append(tweetList,info)
      lim -= 1
    }
  }
  if tweetList == nil {
      tweetList = []Item{}
  }
  return tweetList, nil
}
