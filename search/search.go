package main

import (
    "time"
    "context"
    "log"
    "net/http"
    "github.com/onrik/logrus/filename"
    log "github.com/sirupsen/logrus"
    "encoding/json"
    "github.com/gorilla/mux"
    "github.com/mongodb/mongo-go-driver/mongo"
    "github.com/mongodb/mongo-go-driver/bson"
)

type params struct {
    Timestamp int64 `json:"timestamp,string"`
    Limit int `json:"limit,string"`
    Q string `json:"q,string,omitempty"`
    Un string `json:"username,string,omitempty"`
    Following bool `json:"following,string,omitempty"`
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
    log.AddHook(filename.NewHook())
    log.SetLevel(log.ErrorLevel)
    http.ListenAndServe(":8006", nil)
}

func getUsername(r *http.Request) (string, error) {
    cookie, err := r.Cookie("username")
    if err != nil {
        return "", err
    } else {
        return cookie.Value, nil
    }
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
       log.Error("Could not decode JSON")
       json.NewEncoder(w).Encode(r)
       return
    }
    //Error checking and defaulting the parameters
    if(start.Timestamp == 0){
      start.Timestamp = time.Now().Unix()
    }
    if(start.Limit == 0){
      start.Limit = 25
    }
    if(start.Limit > 100){
      r.Status = "error"
      r.Error = "Limit must be under 100"
      log.Error("Limit exceeded 100")
      json.NewEncoder(w).Encode(r)
    }
    if(start.Following == nil){
      start.Following = true
    }
    //Generating the list of items
    itemList, err := generateList(start, req)
    //Error checking the returned list and returning the proper json response
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

func getFollowingList(string username, Collection c) (string){
  doc := bson.NewDocument(bson.EC.String("username",username))
  user,err := col.Find(context.Background(),doc)
  if err != nil{
    log.Error("Could not find user in DB")
    return []string
  }
  res, err4 := row.Lookup("following")
  if err4 == nil{
    return res.Value().ReaderArray()
  }else{
    log.Error(err4)
    return []string
  }
}

func generateList(sPoint params, r *http.Request) ([]Item, error){
  //Connecting to db and setting up the collection
  client, err := mongo.NewClient("mongodb://localhost:27017")
  if err != nil {
      log.Println("Error Connecting")
      log.Error("Problem connecting to MongoDB")
      return nil, err
  }
  db := client.Database("twitter")
  col := db.Collection("tweets")

  var tweetList []Item
  var info Item
  var prop property
  user,err := getUsername(r)
  doc := bson.NewDocument(bson.EC.SubDocumentFromElements("timestamp",bson.EC.Int64("$lte", (int64)(sPoint.Timestamp),)))
  if(sPoint.Un != nil){
    doc.Append(bson.EC.String("username",sPoint.Un))
  }
  if(sPoint.Following == true && user != ""){
    followingList:=getFollowingList(user,col)
    doc.Append(bson.EC.SubDocumentFromElements("username",bson.EC.String("$in", followingList))
  }else{
    log.Info("No logged in user found")
    return nil, err
  }
  set,err := col.Find(
      context.Background(),
      doc)
  //error checking, if valid then it retrieves the limit's amount of document
  if err != nil {
    log.Println("Error querying")
    log.Error("Problem with query")
      return nil, err
  } else {
    lim := sPoint.Limit
    for set.Next(context.Background()) && lim>0{

      row := bson.NewDocument()
      err = set.Decode(&info)


      //log.Println(row)

      // res, err4 := row.Lookup("content")
      // if err4 == nil{
      //   info.Content = res.Value().StringValue()
      // }
      //
      // res, err4 = row.Lookup("_id")
      // if err4 == nil{
      //   info.ID = res.Value().ObjectID().Hex()
      // }
      //
      // res, err4 = row.Lookup("likes")
      // if err4 == nil{
      //   prop.Likes = (int)(res.Value().Int32())
      //   info.Property = prop
      // }
      //
      // res, err4 = row.Lookup("username")
      // if err4 == nil{
      //   info.Username = res.Value().StringValue()
      // }
      //
      // res, err4 = row.Lookup("retweeted")
      // if err4 == nil{
      //   info.Retweeted = (int)(res.Value().Int32())
      // }
      //
      // res, err4 = row.Lookup("timestamp")
      // if err4 == nil{
      //   info.Timestamp= res.Value().Int64()
      // }

      tweetList = append(tweetList,info)
      lim -= 1
    }
  }
  if tweetList == nil {
      tweetList = []Item{}
  }
  return tweetList, nil
}
