package main

import (
    "time"
    "context"
    "net/http"
    "github.com/onrik/logrus/filename"
    log "github.com/sirupsen/logrus"
    "encoding/json"
    "github.com/gorilla/mux"
    "github.com/mongodb/mongo-go-driver/mongo"
    "TwitterClone/user"
    "github.com/mongodb/mongo-go-driver/bson"
    "github.com/mongodb/mongo-go-driver/bson/objectid"
)

type params struct {
    Timestamp int64 `json:"timestamp,string"`
    Limit int `json:"limit,omitempty"`
    Q string `json:"q,omitempty"`
    Un string `json:"username,omitempty"`
    Following *bool `json:"following,omitempty"`
}

type Item struct {
    ID objectid.ObjectID `json:"-" bson:"_id"`
    IdString string `json:"id" bson:"id"`
    Content string `json:"content" bson:"content"`
    Username string `json:"username" bson:"username"`
    Property property `json:"property" bson:"property"`
    Retweeted int `json:"retweeted" bson:"retweeted"`
    Timestamp int64 `json:"timestamp" bson:"timestamp"`
}
type property struct {
  Likes int `json:"likes"`
}
type res struct {
  Status string `json:"status"`
  Items []Item `json:"items"`
  Error string `json:"error,omitempty"`
}

var client *mongo.Client

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/search", search).Methods("POST")
    http.Handle("/", r)
    log.AddHook(filename.NewHook())
    log.SetLevel(log.InfoLevel)
    var err error
    client, err = mongo.NewClient("mongodb://mongo.db:27017")
    if err != nil {
        log.Fatal("Problem connecting to MongoDB")
    }
    http.ListenAndServe(":8006", nil)
}

func getUsername(r *http.Request) (string, error) {
    cookie, err := r.Cookie("username")
    if err != nil {
        return "", err  //CHANGE THIS
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
      def := new(bool)
      *def = true
      start.Following = def
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

func getFollowingList(username string, db mongo.Database) ([]string){
  doc := bson.NewDocument(bson.EC.String("username",username))
  c := db.Collection("users")
  userFind,err := c.Find(context.Background(),doc)
  if err != nil{
    log.Error("Could not find user in DB")
    return nil
  }

  var foundUser user.User
  userFind.Decode(foundUser)
  log.Info(foundUser)
  return foundUser.Following
}

func generateList(sPoint params, r *http.Request) ([]Item, error){
  //Connecting to db and setting up the collection
  log.Info(sPoint)
  db := client.Database("twitter")
  col := db.Collection("tweets")

  var tweetList []Item
  var info Item
  //var prop property
  user,err := getUsername(r)
  if err != nil{
    log.Error(err)
    return nil,err
  }
  doc := bson.NewDocument(bson.EC.SubDocumentFromElements("timestamp",bson.EC.Int64("$lte", (int64)(sPoint.Timestamp),)))
  if(sPoint.Un != ""){
    doc.Append(bson.EC.String("username",sPoint.Un))
  }
  if(*(sPoint.Following) == true){
    followingList:=getFollowingList(user,*db)
    bArray := bson.NewArray()
    for _,element := range followingList{
      bArray.Append(bson.EC.String("fUsername",element).Value())
    }
    doc.Append(bson.EC.SubDocumentFromElements("username",bson.EC.Array("$in", bArray)))
  }
  log.Info(doc)
  set,err := col.Find(
      context.Background(),
      doc)
  //error checking, if valid then it retrieves the limit's amount of document
  if err != nil {
    log.Error("Problem with query")
      return nil, err
  } else {
    log.Info(set)
    lim := sPoint.Limit
    for set.Next(context.Background()) && lim>0{
      //row := bson.NewDocument()
      err = set.Decode(&info)
      info.IdString = info.ID.Hex()
      tweetList = append(tweetList,info)
      lim -= 1
    }
  }
  if tweetList == nil {
      tweetList = []Item{}
  }
  return tweetList, nil
}
