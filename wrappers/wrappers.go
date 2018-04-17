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
)


func main() {
    log.AddHook(filehook.NewHook())
}
var client *mongo.Client

func NewClient() (*mongo.Client, error) {
    if client != nil {
        return client, nil
    }
    var err error
    var ClientOpt = &mongo.ClientOptions{}
    opts := ClientOpt.
        MaxConnIdleTime(time.Second * 30)
    client, err = mongo.NewClientWithOptions(
        "mongodb://mongo-query-router:27017", opts)
    if err != nil {
        log.Error(err)
    }
    return client, err
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
