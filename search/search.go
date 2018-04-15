package search

import (
    "time"
    "context"
    "net/http"
    "github.com/sirupsen/logrus"
    //"github.com/logrustash"
    "gopkg.in/sohlich/elogrus.v3"
    //"github.com/sohlich/elogrus"
	   "github.com/olivere/elastic"
    "encoding/json"
    "github.com/gorilla/mux"
    "github.com/mongodb/mongo-go-driver/mongo"
    "TwitterClone/user"
    "github.com/mongodb/mongo-go-driver/bson"
    "github.com/mongodb/mongo-go-driver/bson/objectid"
    "TwitterClone/wrappers"
    "TwitterClone/item"
)

type params struct {
    Timestamp int64 `json:"timestamp,string"`
    Limit int `json:"limit,omitempty"`
    Q string `json:"q,omitempty"`
    Un string `json:"username,omitempty"`
    Following *bool `json:"following,omitempty"`
    Rank *string `json:"rank,omitempty"`
    Replies *bool `json:"replies"`
    HasMedia bool `json:"hasMedia"`
    Parent *string `json:"parent"`
}

type property struct {
  Likes int `json:"likes"`
}

type res struct {
  Status string `json:"status"`
  Items []item.Item `json:"items"`
  Error string `json:"error,omitempty"`
}

var log logrus.Logger
func main() {
    r := mux.NewRouter()
    r.HandleFunc("/search", SearchHandler).Methods("POST")
    http.Handle("/", r)

    log := logrus.New()
	client, err := elastic.NewClient(elastic.SetURL("http://localhost:9200"))
	if err != nil {
        log.Panic(err)
	}
	hook, err := elogrus.NewAsyncElasticHook(client, "localhost", logrus.DebugLevel, "twiti")
	if err != nil {
		log.Panic(err)
	}
	log.Hooks.Add(hook)
    log.WithFields(logrus.Fields{
		"name": "joe",
        "age":  42,
    }).Info("Hello world!")
    // Log to a file
    // var f *os.File
    // var err error
    // log, f, err = wrappers.FileLogger("search.log", os.O_CREATE | os.O_RDWR,
    //     0666)
    // if err != nil {
    //     log.Fatal("Logging file could not be opened.")
    // }
    // f.Truncate(0)
    // f.Seek(0, 0)
    // defer f.Close()
    log.SetLevel(logrus.InfoLevel)
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

func SearchHandler(w http.ResponseWriter, req *http.Request) {
    startTime := time.Now()
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
    if(start.Rank == nil){
      def := new(string)
      *def = "interest"
      start.Rank = def
    }
    if(start.Replies == nil){
      def := new(bool)
      *def = true
      start.Replies = def
    }
    log.WithFields(logrus.Fields{"timestamp": start.Timestamp, "limit": start.Limit,
    "Q": start.Q, "un": start.Un, "following": *start.Following}).Info("params")
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
    elapsed := time.Since(startTime)
    log.Info("elapsed: " + elapsed.String())
  json.NewEncoder(w).Encode(r)
}

func getFollowingList(username string, db mongo.Database) ([]string){
  filter := bson.NewDocument(bson.EC.String("username",username))
  c := db.Collection("users")
  var foundUser user.User
  err := c.FindOne(context.Background(), filter).Decode(&foundUser)
  if err != nil{
    log.Error("Could not find user in DB")
    return nil
  }
  log.Info(foundUser)
  return foundUser.Following

}

func generateList(sPoint params, r *http.Request) ([]item.Item, error){
  //Connecting to db and setting up the collection
  client, err := wrappers.NewClient()
  if err != nil {
    log.Error("Could not connect to Mongo.")
    return nil, err
  }
  log.Info(sPoint)
  db := client.Database("twitter")
  col := db.Collection("tweets")

  var tweetList []item.Item
  var info item.Item
  //var prop property
  user,err := getUsername(r)
  if err != nil{
    log.Error(err)
    return nil,err
  }
  doc := bson.NewArray(bson.VC.DocumentFromElements(bson.EC.SubDocumentFromElements("$match",bson.EC.SubDocumentFromElements("timestamp",bson.EC.Int64("$lte", (int64)(sPoint.Timestamp)),),),))
  if(sPoint.Un != ""){
    doc.Append(bson.VC.DocumentFromElements(bson.EC.SubDocumentFromElements("$match",bson.EC.String("username",sPoint.Un))))
  }
  if(*(sPoint.Following) == true){
    followingList:=getFollowingList(user,*db)
    bArray := bson.NewArray()
    for _,element := range followingList{
      bArray.Append(bson.EC.String("fUsername",element).Value())
    }
    log.Info(bArray)
    doc.Append(bson.VC.DocumentFromElements(bson.EC.SubDocumentFromElements("$match",bson.EC.SubDocumentFromElements("username",bson.EC.Array("$in", bArray)))))
  }
  if(sPoint.Q != ""){
      doc.Append(bson.VC.DocumentFromElements(bson.EC.SubDocumentFromElements("$match",bson.EC.Regex("content", sPoint.Q, ""))))
  }
  if(*(sPoint.Rank) == "interest"){
    log.Info("Interest is the ranking")
    doc.Append(bson.VC.DocumentFromElements(bson.EC.SubDocumentFromElements("$sort",bson.EC.Int32("property.likes", -1),bson.EC.Int32("retweeted", -1))))
  }

  if(*(sPoint.Replies) == false){
    //exclude reply tweets
    doc.Append(bson.VC.DocumentFromElements(bson.EC.SubDocumentFromElements("$match",bson.EC.SubDocumentFromElements("username",bson.EC.Regex("$not", "reply","")))))

  }
  if(sPoint.Parent != nil){
    //only return tweets where parent = given parentId
    poid,err := objectid.FromHex(*(sPoint.Parent))
    if err != nil {
        log.Error("Invalid Parent ID")
        return nil, err
    }
    doc.Append(bson.VC.DocumentFromElements(bson.EC.SubDocumentFromElements("$match",bson.EC.ObjectID("parent",poid))))
  }
  log.Info(doc)
  set,err := col.Aggregate(
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
      if(sPoint.HasMedia == true){
        //only query tweets with Media
        if(info.MediaIDs != nil){
          tweetList = append(tweetList,info)
          lim -= 1
        }else{
          continue;
        }
      }else{
        tweetList = append(tweetList,info)
        lim -= 1
      }
    }
  }
  if tweetList == nil {
      tweetList = []item.Item{}
  }
  return tweetList, nil
}
