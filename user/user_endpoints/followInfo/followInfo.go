package followInfo

import (
    "context"
    "net/http"
    "errors"
    logrus "github.com/sirupsen/logrus"
    "encoding/json"
    "TwitterClone/wrappers"
    "time"
    "github.com/gorilla/mux"
    "github.com/mongodb/mongo-go-driver/bson"
	  "github.com/mongodb/mongo-go-driver/mongo"
    "strconv"
)

type followList struct {
  Following []string `bson:"following,omitempty"`
  Followers []string `bson:"followers,omitempty"`
}

type response struct {
    Status string `json:"status"`
    Users []string `json:"users"`
    Error string `json:"error,omitempty"`
}
var Log *logrus.Logger
func main() {
    Log.SetLevel(logrus.ErrorLevel)
}

func encodeResponse(w http.ResponseWriter, response interface{}) error {
    return json.NewEncoder(w).Encode(response)
}

func getUsername(r *http.Request) (string){
  vars := mux.Vars(r)
  username := vars["username"]
  Log.Debug(username)
  return username
}

func checkLimit(r *http.Request) (int64,error){
  limit, err := strconv.ParseInt(r.URL.Query().Get("limit"),10,64)
  if err != nil{
      limit = 50
  }else if limit != 0 && limit > 200{
    Log.Error("Limit exceeds 200")
    return 0,errors.New("Limit exceeds 200")
  }
  Log.Info(limit)
  return limit,nil
}

func GetFollowingHandler(w http.ResponseWriter, r *http.Request) {
start := time.Now()
    var resp response
    lim, e := checkLimit(r)
    if e != nil{
      Log.Info(e)
      resp.Status = "error"
      resp.Error = e.Error()
      encodeResponse(w,resp)
    }
    username := getUsername(r)
    list, err := findUserFollowing(username,lim)
    if err != nil {
        Log.Info(err)
        resp.Status = "error"
        resp.Error = err.Error()

    elapsed := time.Since(start)
    Log.Info("Get Following elapsed: " + elapsed.String())
        encodeResponse(w,resp)
    }else{
      resp.Status = "OK"
      resp.Users = list

    elapsed := time.Since(start)
    Log.Info("Get Following elapsed: " + elapsed.String())
      encodeResponse(w, resp)
    }
}

func GetFollowersHandler(w http.ResponseWriter, r *http.Request) {
start := time.Now()
    var resp response
    lim, e := checkLimit(r)
    if e != nil{
      Log.Info(e)
      resp.Status = "error"
      resp.Error = e.Error()
      encodeResponse(w,resp)
    }
    username := getUsername(r)
    list, err := findUserFollowers(username,lim)
    if err != nil {
        Log.Info(err)
        resp.Status = "error"
        resp.Error = err.Error()

    elapsed := time.Since(start)
    Log.Info("Get Followers elapsed: " + elapsed.String())
        encodeResponse(w,resp)
    }else{
      resp.Status = "OK"
      resp.Users = list

    elapsed := time.Since(start)
    Log.Info("Get Followers elapsed: " + elapsed.String())
      encodeResponse(w, resp)
    }
}

func findUserFollowing(username string, lim int64) ([]string, error) {
    following,err := findUserFollow(username,"following", lim)
    list := []string{}
    if err != nil{
      Log.Error(err)
      return nil,errors.New(err.Error())
    }
    if following != nil{
      return following, nil
    }else{
      return list, nil
    }
}

func findUserFollowers(username string, lim int64) ([]string, error) {
    followers,err := findUserFollow(username, "followers", lim)
    list := []string{}
    if err != nil{
      Log.Error(err)
      return nil,errors.New(err.Error())
    }

    if followers != nil{
      return followers,nil
    }else{
      return list, nil
    }
}

func findUserFollow(username string, follow string, lim int64) ([]string,error){
  dbStart := time.Now()
  client, err := wrappers.NewClient()
  if err != nil {
      return nil,err
  }
  db := client.Database("twitter")
  coll := db.Collection("users")
  filter := bson.NewDocument(bson.EC.String("username", username))
  proj := bson.NewDocument(bson.EC.SubDocumentFromElements(follow,bson.EC.Int64("$slice",lim)), bson.EC.Int32("_id",0))

  var fArray followList
  option, err := mongo.Opt.Projection(proj)
  if err != nil {
      return nil,err
  }

  Log.Info(option)
  err = coll.FindOne(context.Background(),
      filter,option).Decode(&fArray)
  Log.Info(fArray)
  if err != nil{
    elapsed := time.Since(dbStart)
    Log.WithFields(logrus.Fields{"msg":"Check if user exists time elapsed", "timeElapsed":elapsed.String()}).Error("Could not find user")
    Log.Error(err)
    return nil,errors.New("Could not find user")
  }

  elapsed := time.Since(dbStart)
  Log.WithFields(logrus.Fields{"msg":"Check if user exists time elapsed", "timeElapsed":elapsed.String()}).Info()
  if(follow == "followers"){
    return fArray.Followers,nil
  }else{
    return fArray.Following,nil
  }
}
