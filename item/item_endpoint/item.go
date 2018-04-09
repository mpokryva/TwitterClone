package main

import (
    "context"
    "errors"
    "github.com/onrik/logrus/filename"
    log "github.com/sirupsen/logrus"
    "net/http"
    "encoding/json"
    "github.com/gorilla/mux"
    "github.com/mongodb/mongo-go-driver/bson"
    "github.com/mongodb/mongo-go-driver/mongo"
    "github.com/mongodb/mongo-go-driver/bson/objectid"
    "TwitterClone/wrappers"
    "TwitterClone/item"
)


type response struct {
    Status string `json:"status"`
    Item item.Item `json:"item,omitempty"`
    Error string `json:"error,omitempty"`
}

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/item/{id}", getItemHandler).Methods("GET")
    //r.HandleFunc("/item/{id}/like", likeItemHandler).Methods("POST")
    r.HandleFunc("/item/{id}", deleteItemHandler).Methods("DELETE")
    log.AddHook(filename.NewHook())
    log.SetLevel(log.DebugLevel)
    http.Handle("/", r)
    log.Fatal(http.ListenAndServe(":8005", nil))
}

//LIKE ITEM FUNCTIONS START HERE
/*
func likeItemHandler(w http.ResponseWriter, r *http.Request) {
    id := mux.Vars(r)["id"]
    log.Debug(id)

    var res response
    username, err := checkLogin(r)
    if err != nil {
      log.Error("User not logged in")
        res.Status = "error"
        res.Error = "User not logged in."
    } else {
        like, err := decodeRequest(r)
        if (err != nil) {
          log.Error("JSON decoding error")
            res.Status = "error"
            res.Error = "JSON decoding error."
        } else {
            //res = likeItemEndpoint(id, username, like)
            
        }
    }
    encodeResponse(w, res)
}
*/
func decodeRequest(r *http.Request) (bool, error) {
    decoder := json.NewDecoder(r.Body)
    var like bool;
    err := decoder.Decode(&like)
    return like, err
}
/*
func likeItemEndpoint(id string, username string, like bool) response {
    return likeItem(id, username)
}*/
/*
func likeItem(id string, username string, like bool) response {
    var resp response
    client, err := wrappers.NewClient()
    db := client.Database("twitter")
    col := db.Collection("tweets")

    if err != nil {
        log.Error("Error connecting to database")
        resp.Status = "error"
        resp.Error = "Database unavailable"
        return resp
    }

    objectid,err := objectid.FromHex(id)
    if err != nil {
        resp.Status = "error"
        resp.Error = "Invalid Item ID format"
        return resp
    }

	// Check if user to like exists.
    // Assuming that logged in user exists (not bogus cookie).
    checkUserFilter := bson.NewDocument(
        bson.EC.String("username", username))
    err = col.FindOne(context.Background(), checkUserFilter)
    if err != nil {
        log.Info(err)
        return errors.New("User doesn't exist.")
    }
    // Are we incrementing or decrementing the number of likes?
    var listOp string
    var countInc int32
    if like {
        listOp = "$addToSet"
        countInc = 1
    } else {
        listOp = "$pull"
        countInc = -1
    }

    // Update like count.
    filter := bson.NewDocument(bson.EC.ObjectID("_id", objectid))
    update := bson.NewDocument(bson.EC.SubDocumentFromElements("$inc",
    bson.EC.Int32("likes", countInc)))
    err = UpdateOne(coll, filter, update)
    if err != nil {
        log.Error("Did not find ObjectID")
        resp.Status = "error"
        resp.Error = "Invalid Item ID"
        return resp
    }
    resp.Status = "OK"
    resp.Error = ""
    // log.Debug("Encoded!")
    return resp
}
*/
func UpdateOne(coll *mongo.Collection, filter interface{}, update interface{}) error {
    result, err := coll.UpdateOne(
        context.Background(),
        filter, update)
    var success = false
    if result != nil {
        log.Debug(*result)
        success = result.ModifiedCount == 1
    }
    if err != nil {
        return err
    } else if !success {
        return errors.New("Database is operating normally, but like update " +
        "operation failed.")
    } else {
        return nil;
    }
}

//LIKE ITEM ENDPOINT ENDS HERE

//GET ITEM FUNCTIONS START HERE
func getItemHandler(w http.ResponseWriter, r *http.Request) {
    var res response
    id := mux.Vars(r)["id"]
    log.Debug(id)
    res = getItemEndpoint(id)
    encodeResponse(w, res)
}

func getItemEndpoint(id string) response {
    return getItem(id)
}

func getItem(id string) response {
    var item item.Item
    var resp response
    client, err := wrappers.NewClient()
    if err != nil {
        log.Error(err)
        resp.Status = "error"
        resp.Error = err.Error()
        return resp
    }
    db := client.Database("twitter")
    col := db.Collection("tweets")
    objectid, err := objectid.FromHex(id)
    log.Debug(objectid)

    if err != nil {
        log.Error(err)
        resp.Status = "error"
        resp.Error = err.Error()
        return resp
    }
    filter := bson.NewDocument(bson.EC.ObjectID("_id", objectid))
    err = col.FindOne(
        context.Background(),
        filter).Decode(&item)
    if err != nil {
        log.Error(err)
        resp.Status = "error"
        resp.Error = err.Error()
        return resp
    }
    /*
    for item.Next(context.Background()){
        row := bson.NewDocument()
        err = item.Decode(row)

        res, err4 := row.Lookup("content")
        if err4 == nil{
          info.Content = res.Value().StringValue()
        }
        res, err4 = row.Lookup("_id")
        if err4 == nil{
          info.ID = res.Value().ObjectID().Hex()
        }
        res, err4 = row.Lookup("username")
        if err4 == nil{
          info.Username = res.Value().StringValue()
        }
        res, err4 = row.Lookup("likes")
        if err4 == nil{
          prop.Likes = (int)(res.Value().Int32())
          info.Property = prop
        }
        res, err4 = row.Lookup("retweeted")
        if err4 == nil{
          info.Retweeted = (int)(res.Value().Int32())
        }
        res, err4 = row.Lookup("timestamp")
        if err4 == nil{
          info.Timestamp= res.Value().Int64()
        }
        res, err4 = row.Lookup("childType")
        if err4 == nil{
          info.Timestamp= res.Value().StringValue()
        }
        res, err4 = row.Lookup("parent")
        if err4 == nil{
          info.Timestamp= res.Value().ObjectID().Hex()
        }
        res, err4 = row.Lookup("media")
        if err4 == nil{
        	//UPDATE THIS TO RETRIEVE THE MEDIA ID's
          info.Timestamp= res.Value().Int64()
        }
    }
    */
    resp.Status = "OK"
    resp.Item = item
    return resp
}
//GET ITEM FUNCTIONS END HERE

//DELETE ITEM FUNCTIONS START HERE
func deleteItemHandler(w http.ResponseWriter, r *http.Request) {
    var res response
    id := mux.Vars(r)["id"]
    log.Debug(id)
    res = deleteItemEndpoint(id)
    encodeResponse(w, res)
}

func deleteItemEndpoint(id string) response {
    return deleteItem(id)
}

func deleteItem(id string) response {
    var resp response 
    client, err := wrappers.NewClient()
    if err != nil {
        log.Error("Error connecting to database")
        resp.Status = "error"
        resp.Error = "Database unavailable"
        return resp
    }
    db := client.Database("twitter")
    col := db.Collection("tweets")
    objectid,err := objectid.FromHex(id)
    if err != nil {
        resp.Status = "error"
        resp.Error = "Invalid Item ID"
        return resp
    }

    doc := bson.NewDocument(bson.EC.ObjectID("_id", objectid))
    _, err = col.DeleteOne(
        context.Background(),
        doc)
    if err != nil {
        log.Error("Did not find ObjectID")
        resp.Status = "error"
        resp.Error = "Invalid Item ID"
        return resp
    }
    resp.Status = "OK"
    resp.Error = ""
    log.Debug("Encoded!")
    return resp
}
//DELETE ITEM FUNCTIONS END HERE

//MISC
func encodeResponse(w http.ResponseWriter, response interface{}) error {
    return json.NewEncoder(w).Encode(response)
}

func validateItem(it item.Item) bool {
    valid := true
    return valid
}

func checkLogin(r *http.Request) (string, error) {
    cookie, err := r.Cookie("username")
    if err != nil {
        return "", err
    } else {
        return cookie.Value, nil
    }
}
