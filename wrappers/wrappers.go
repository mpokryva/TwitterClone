package wrappers

import (
    "github.com/mongodb/mongo-go-driver/mongo"
    //"github.com/mongodb/mongo-go-driver/bson"
    //"github.com/mongodb/mongo-go-driver/bson/objectid"
    "github.com/onrik/logrus/filename"
    log "github.com/sirupsen/logrus"
)

func main() {
    log.AddHook(filename.NewHook())
}

func NewClient() (*mongo.Client, error) {
    client, err := mongo.NewClient("mongodb://mongo.db:27017")
    if (err != nil) {
        log.Error(err)
    }
    return client, err
}
