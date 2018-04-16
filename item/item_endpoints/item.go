package item_endpoints

import (
    "context"
    "errors"
    logrus "github.com/sirupsen/logrus"
    "net/http"
    "os"
    "encoding/json"
    "github.com/gorilla/mux"
    "github.com/mongodb/mongo-go-driver/bson"
    "github.com/mongodb/mongo-go-driver/mongo"
    "github.com/mongodb/mongo-go-driver/bson/objectid"
    "TwitterClone/wrappers"
    "TwitterClone/item"
)

type Like struct {
    ID objectid.ObjectID `json:"id" bson:"_id"`
    Username string `json:"username" bson:"username"`
}

type response struct {
    Status string `json:"status"`
    Item item.Item `json:"item,omitempty"`
    Error string `json:"error,omitempty"`
}

//response for Like
type responseL struct {
    Status string `json:"status"`
    Error string `json:"error,omitempty"`
}

//post params for like
type Req struct {
    Like *bool `json:"like"`
}
var Log *logrus.Logger
func main() {
    r := mux.NewRouter()
    r.HandleFunc("/item/{id}", GetItemHandler).Methods("GET")
    r.HandleFunc("/item/{id}/like", LikeItemHandler).Methods("POST")
    r.HandleFunc("/item/{id}", DeleteItemHandler).Methods("DELETE")
    // Log to a file
    var f *os.File
    var err error
    log, f, err = wrappers.FileLogger("item.log", os.O_CREATE | os.O_RDWR,
        0666)
    if err != nil {
        Log.Fatal("Logging file could not be opened.")
    }
    f.Truncate(0)
    f.Seek(0, 0)
    defer f.Close()
    Log.SetLevel(logrus.ErrorLevel)
    http.Handle("/", r)
    Log.Fatal(http.ListenAndServe(":8005", nil))
}

//LIKE ITEM FUNCTIONS START HERE

func LikeItemHandler(w http.ResponseWriter, r *http.Request) {
    id := mux.Vars(r)["id"]
    Log.Debug(id)

    var res responseL
    username, err := checkLogin(r)
    if err != nil {
      Log.Error("User not logged in")
        res.Status = "error"
        res.Error = "User not logged in."
    } else {
        req, err := decodeRequest(r)
        if (err != nil) {
          Log.Error(r)
          Log.Error("JSON decoding error")
            res.Status = "error"
            res.Error = "JSON decoding error."
        } else {
            res = likeItemEndpoint(id, username, *req.Like)
        }
    }
    encodeResponse(w, res)
}

func decodeRequest(r *http.Request) (Req, error) {
    decoder := json.NewDecoder(r.Body)
    var like Req
    err := decoder.Decode(&like)
    return like, err
}

func likeItemEndpoint(id string, username string, like bool) responseL {
    return likeItem(id, username, like)
}

func likeItem(id string, username string, like bool) responseL {
    var resp responseL
    client, err := wrappers.NewClient()
    db := client.Database("twitter")
    // col := db.Collection("users")

    // if err != nil {
    //     Log.Error("Error connecting to database")
    //     resp.Status = "error"
    //     resp.Error = "Database unavailable"
    //     return resp
    // }

    objectid,err := objectid.FromHex(id)
    if err != nil {
        resp.Status = "error"
        resp.Error = "Invalid Item ID format"
        return resp
    }

	// // Check if user liking exists.
 //    // Assuming that logged in user exists (not bogus cookie).
 //    checkUserFilter := bson.NewDocument(
 //        bson.EC.String("username", username))
 //    err = col.FindOne(context.Background(), checkUserFilter)
 //    if err != nil {
 //        Log.Info("User does not exist")
 //        resp.Status = "error"
 //        resp.Error = "User in cookie does not exist"
 //        return resp
 //    }

    col := db.Collection("likes")
    if err != nil {
        Log.Error("Error connecting to database")
        resp.Status = "error"
        resp.Error = "Database unavailable"
        return resp//e object into db (username, itemid) *IF NOT EXISTS maybe*
    //////////////////
    //THIS IS BROKEN
    }
    if like {
    // Log.Debug("We are liking")
    var likeItem Like
    likeItem.ID = objectid
    likeItem.Username = username
    //insert lik
    //////////////////

    _, err = col.InsertOne(context.Background(), &likeItem)
        if err != nil {
            Log.Info("Error adding to likes collection")
            resp.Status = "error"
            resp.Error = "Error liking please try again"
            return resp
        }
    }else{
    //delete like object from db (username, itemid)
    var likeItem Like
    likeItem.ID = objectid
    likeItem.Username = username
    _, err = col.DeleteOne(context.Background(), &likeItem)
        if err != nil {
            Log.Info("User does not have an entry in likes for this itemid")
            resp.Status = "error"
            resp.Error = "You have not liked this tweet before"
            return resp
        }
    }
    col = db.Collection("tweets")

    // Are we incrementing or decrementing the number of likes?
    var countInc int32
    if like {
        countInc = 1
    } else {
        countInc = -1
    }

    // Update like count.
    //////////////////
    //THIS IS BROKEN or MAYBE I PUT THE WRONG OBJECTID IN POSTMAN??
    //i think my query is wwrong..
    //http://130.245.170.tem/5acd9afc5b97328b39569268/like
    //{"status":"error","error":"Invalid Item ID"}
    //////////////////
    filter := bson.NewDocument(bson.EC.ObjectID("_id", objectid))
    update := bson.NewDocument(bson.EC.SubDocumentFromElements("$inc",
        bson.EC.Int32("property.likes", countInc)))
    err = UpdateOne(col, filter, update)
    if err != nil {
        Log.Error("Did not find ObjectID")
        resp.Status = "error"
        resp.Error = "Invalid 65/iItem ID"
        return resp
    }
    resp.Status = "OK"
    resp.Error = ""
    // Log.Debug("Encoded!")
    return resp
}

func UpdateOne(coll *mongo.Collection, filter interface{}, update interface{}) error {
    result, err := coll.UpdateOne(
        context.Background(),
        filter, update)
    var success = false
    if result != nil {
        Log.Debug(*result)
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
func GetItemHandler(w http.ResponseWriter, r *http.Request) {
    var res response
    id := mux.Vars(r)["id"]
    Log.Debug(id)
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
        Log.Error(err)
        resp.Status = "error"
        resp.Error = err.Error()
        return resp
    }
    db := client.Database("twitter")
    col := db.Collection("tweets")
    objectid, err := objectid.FromHex(id)
    Log.Debug(objectid)

    if err != nil {
        Log.Error(err)
        resp.Status = "error"
        resp.Error = err.Error()
        return resp
    }
    filter := bson.NewDocument(bson.EC.ObjectID("_id", objectid))
    err = col.FindOne(
        context.Background(),
        filter).Decode(&item)
    if err != nil {
        Log.Error(err)
        resp.Status = "error"
        resp.Error = err.Error()
        return resp
    }
    resp.Status = "OK"
    resp.Item = item
    return resp
}
//GET ITEM FUNCTIONS END HERE

//DELETE ITEM FUNCTIONS START HERE
func DeleteItemHandler(w http.ResponseWriter, r *http.Request) {
    var statusCode int
    id := mux.Vars(r)["id"]
    Log.Debug(id)
    statusCode = deleteItemEndpoint(id)
    w.WriteHeader(statusCode)
}

func deleteItemEndpoint(id string) int {
    return deleteItem(id)
}

func deleteItem(id string) int {
    client, err := wrappers.NewClient()
    if err != nil {
        Log.Error(err)
        return http.StatusInternalServerError
    }
    db := client.Database("twitter")
    col := db.Collection("tweets")
    objectid, err := objectid.FromHex(id)
    if err != nil {
        Log.Error(err)
        return http.StatusBadRequest
    }
    // Pull item from database.
    var it item.Item
    doc := bson.NewDocument(bson.EC.ObjectID("_id", objectid))
    err = col.FindOne(context.Background(), doc).Decode(&it)
    if err != nil {
        Log.Info("item does not exist")
        Log.Error(err)
        return http.StatusBadRequest
    }

    // Delete associated media, if it exists.
    var result *mongo.UpdateResult
    if it.MediaIDs != nil {
        col = db.Collection("media")
        bArray := bson.NewArray()
        for _, mOID := range it.MediaIDs {
            bArray.Append(bson.VC.ObjectID(mOID))
        }
        filter := bson.NewDocument(
            bson.EC.SubDocumentFromElements("_id",
            bson.EC.Array("$in", bArray)))
        update := bson.NewDocument(
            bson.EC.SubDocumentFromElements("$pull",
            bson.EC.ObjectID("item_ids", it.ID)))
            result, err = col.UpdateMany(context.Background(), filter, update)
            Log.Debug(result)
    }
    if err != nil {
        Log.Error(err)
        return http.StatusInternalServerError
    }
    // Successfully deleted media ids.
    // Delete actual item.
    doc = bson.NewDocument(bson.EC.ObjectID("_id", objectid))
    _, err = col.DeleteOne(
        context.Background(),
        doc)
    if err != nil {
        Log.Error("Did not find item when deleting.")
        Log.Error(err)
        return http.StatusInternalServerError
    }
    return http.StatusOK
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
