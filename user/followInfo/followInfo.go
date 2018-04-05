package main

import (
    "context"
    "net/http"
    "errors"
    //"time"
    "github.com/onrik/logrus/filename"
    log "github.com/sirupsen/logrus"
    "encoding/json"
    "TwitterClone/wrappers"
    "github.com/gorilla/mux"
    //"github.com/mongodb/mongo-go-driver/mongo"
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

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/user/{username}/following", followingHandler).Methods("GET")
    r.HandleFunc("/user/{username}/followers", followersHandler).Methods("GET")
    http.Handle("/", r)
    log.AddHook(filename.NewHook())
    log.SetLevel(log.DebugLevel)
    log.Fatal(http.ListenAndServe(":8008", nil))
}

func encodeResponse(w http.ResponseWriter, response interface{}) error {
    return json.NewEncoder(w).Encode(response)
}

func getUsername(r *http.Request) (string){
  vars := mux.Vars(r)
  username := vars["username"]
  log.Debug(username)
  return username
}

func checkLimit() (params,error){
  var p params
  if(p.Limit != 0 && p.Limit > 200){
    log.Error("Limit exceeds 200")
    return p,errors.New("Limit exceeds 200")
  }else{
    p.Limit = 50
  }
  return p,nil
}

func followingHandler(w http.ResponseWriter, r *http.Request) {
    var resp response
    p, e := checkLimit()
    if e != nil{
      log.Info(e)
      resp.Status = "error"
      resp.Error = e.Error()
      encodeResponse(w,resp)
    }
    username := getUsername(r)
    list, err := findUserFollowing(username,p)
    if err != nil {
        log.Info(err)
        resp.Status = "error"
        resp.Error = err.Error()
        encodeResponse(w,resp)
    }else{
      resp.Status = "OK"
      resp.Users = list
      encodeResponse(w, resp)
    }
}

func followersHandler(w http.ResponseWriter, r *http.Request) {
    var resp response
    p, e := checkLimit()
    if e != nil{
      log.Info(e)
      resp.Status = "error"
      resp.Error = e.Error()
      encodeResponse(w,resp)
    }
    username := getUsername(r)
    list, err := findUserFollowers(username,p)
    if err != nil {
        log.Info(err)
        resp.Status = "error"
        resp.Error = err.Error()
        encodeResponse(w,resp)
    }else{
      resp.Status = "OK"
      resp.Users = list
      encodeResponse(w, resp)
    }
}

func findUserFollowing(username string, p params) ([]string, error) {
    user,err := findUser(username)
    list := []string{}
    if err != nil{
      log.Error(err)
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
      log.Error(err)
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
  log.Debug(res)
  ra := res.Value().ReaderArray()
  log.Debug(ra)
  for limit < (uint)(p.Limit) {
    result,err := ra.ElementAt(limit)
    if err != nil{
      log.Error(err)
      return nil,errors.New(err.Error())
    }
    list = append(list,result.Value().StringValue())
    limit += 1
  }

  return list,nil
}

func findUser(username string) (user.User,error){
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
  log.Debug(foundUser)
  if err != nil{
    log.Error("Could not find user")
    return foundUser,errors.New("Could not find user")
  }
  return foundUser,nil
}
