package wrappers

import (
    "time"
    "os"
    "github.com/mongodb/mongo-go-driver/mongo"
    //"github.com/mongodb/mongo-go-driver/bson"
    //"github.com/mongodb/mongo-go-driver/bson/objectid"
    filehook "github.com/onrik/logrus/filename"
    log "github.com/sirupsen/logrus"
    "gopkg.in/sohlich/elogrus.v3"
    "github.com/olivere/elastic"
    "encoding/json"
    "TwitterClone/memcached"
    "net/http"
    "bytes"
)


func main() {
    log.AddHook(filehook.NewHook())
}

var mongoClient *mongo.Client
var Log *log.Logger

func NewClient() (*mongo.Client, error) {
    if mongoClient != nil {
        return mongoClient, nil
    }
    var err error
    var ClientOpt = &mongo.ClientOptions{}
    opts := ClientOpt.
    MaxConnIdleTime(time.Second * 30)
    mongoClient, err = mongo.NewClientWithOptions(
        "mongodb://mongo-query-router:27017", opts)
        if err != nil {
            log.Error(err)
        }
        return mongoClient, err
}


func GetMemcached(key string, v interface{}) error {
    // TODO: Change this to proper ip address.
    resp, err := http.Get("http://localhost/memcached/" + key)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    var getRes memcached.GetResponse
    err = json.NewDecoder(resp.Body).Decode(&getRes)
    if err != nil {
        return err
    }
    err = json.Unmarshal(getRes.Value, v)
    return err
}

func SetMemcached(key string, v interface{}) (memcached.SetResponse, error) {
    var setRes memcached.SetResponse
    b, err := json.Marshal(v)
    if err != nil {
        return setRes, err
    }
    Log.Debug(b)
    var setReq memcached.SetRequest
    setReq.Key = key
    setReq.Value = b
    b, err = json.Marshal(&setReq)
    if err != nil {
        return setRes, err
    }
    reqBuf := bytes.NewBuffer(b)
    Log.Debug(reqBuf)
    // TODO: change this to proper ip address.
    resp, err := http.Post("http://localhost/memcached", "application/json", reqBuf)
    if err != nil {
        return setRes, err
    }
    defer resp.Body.Close()
    err = json.NewDecoder(resp.Body).Decode(&setRes)
    return setRes, err
}

func FileElasticLogger (filename string, flag int,
    perm os.FileMode)(*log.Logger, *os.File, error) {
    var logger = log.New()
    logger.AddHook(filehook.NewHook())
    //log to elasticsearch
    client, err := elastic.NewClient(elastic.SetURL("http://localhost:9200"))
    if err != nil {
        logger.Panic(err)
    }
    hook, err := elogrus.NewAsyncElasticHook(client, "localhost", log.InfoLevel, "twiti")
    if err != nil {
        logger.Panic(err)
    }
    logger.Hooks.Add(hook)
    //Log to a file
    f, err := os.OpenFile(filename, flag, perm)
    if err != nil {
        return nil, nil, err
    }
    // Caller should truncate if neeeded.
    logger.Formatter = &log.JSONFormatter{}
    logger.Out = f
    return logger, f, nil
}

func FileLogger (filename string, flag int,
    perm os.FileMode)(*log.Logger, *os.File, error) {
    var logger = log.New()
    logger.AddHook(filehook.NewHook())
    // Log to a file
    f, err := os.OpenFile(filename, flag, perm)
    if err != nil {
        return nil, nil, err
    }
    // Caller should truncate if neeeded.
    logger.Formatter = &log.JSONFormatter{}
    logger.Out = f
    return logger, f, nil
}

// func HookElastic() (*log.Logger){
//   var handlerLog = log.New()
//   	client, err := elastic.NewClient(elastic.SetURL("http://localhost:9200"))
//   	if err != nil {
//           handlerLog.Panic(err)
//   	}
//   	hook, err := elogrus.NewAsyncElasticHook(client, "localhost", log.InfoLevel, "twiti")
//   	if err != nil {
//   		handlerLog.Panic(err)
//   	}
//   	handlerLog.Hooks.Add(hook)
//     return handlerLog
// }
