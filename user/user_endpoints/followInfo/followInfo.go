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
    "TwitterClone/user"
)

type params struct {
  Limit int `json:"limit,string,omitempty"`
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

func checkLimit() (params,error){
  var p params
  if(p.Limit != 0 && p.Limit > 200){
    Log.Error("Limit exceeds 200")
    return p,errors.New("Limit exceeds 200")
  }else{
    p.Limit = 50
  }
  return p,nil
}

func GetFollowingHandler(w http.ResponseWriter, r *http.Request) {
start := time.Now()
    var resp response
    p, e := checkLimit()
    if e != nil{
      Log.Info(e)
      resp.Status = "error"
      resp.Error = e.Error()
      encodeResponse(w,resp)
    }
    username := getUsername(r)
    list, err := findUserFollowing(username,p)
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
    p, e := checkLimit()
    if e != nil{
      Log.Info(e)
      resp.Status = "error"
      resp.Error = e.Error()
      encodeResponse(w,resp)
    }
    username := getUsername(r)
    list, err := findUserFollowers(username,p)
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

func findUserFollowing(username string, p params) ([]string, error) {
    user,err := findUser(username)
    list := []string{}
    if err != nil{
      Log.Error(err)
      return nil,errors.New(err.Error())
    }
    if user.Followers != nil{
      return user.Followers, nil
    }else{
      return list, nil
    }
}

func findUserFollowers(username string, p params) ([]string, error) {
    user,err := findUser(username)
    list := []string{}
    if err != nil{
      Log.Error(err)
      return nil,errors.New(err.Error())
    }
    if user.Followers != nil{
      return user.Followers, nil
    }else{
      return list, nil
    }
}

func createList(p params, res bson.Element) ([]string,error) {
  var list []string
  var limit uint
  limit = 0
  Log.Debug(res)
  ra := res.Value().ReaderArray()
  Log.Debug(ra)
  for limit < (uint)(p.Limit) {
    result,err := ra.ElementAt(limit)
    if err != nil{
      Log.Error(err)
      return nil,errors.New(err.Error())
    }
    list = append(list,result.Value().StringValue())
    limit += 1
  }

  return list,nil
}

func findUser(username string) (user.User,error){
  dbStart := time.Now()
  var foundUser user.User
  client, err := wrappers.NewClient()
  if err != nil {
      return foundUser,err
  }
  db := client.Database("twitter")
  coll := db.Collection("users")
  filter := bson.NewDocument(bson.EC.String("username", username))

  err = coll.FindOne(context.Background(),
      filter).Decode(&foundUser)
  Log.Debug(foundUser)
  if err != nil{
    elapsed := time.Since(dbStart)
    Log.WithFields(logrus.Fields{"msg":"Check if user exists time elapsed", "timeElapsed":elapsed.String()}).Error("Could not find user")
    return foundUser,errors.New("Could not find user")
  }

  elapsed := time.Since(dbStart)
  Log.WithFields(logrus.Fields{"msg":"Check if user exists time elapsed", "timeElapsed":elapsed.String()}).Info()
  return foundUser,nil
}
