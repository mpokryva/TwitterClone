package item

import (
    "github.com/mongodb/mongo-go-driver/bson/objectid"
)

type Item struct {
    ID objectid.ObjectID `json:"id" bson:"_id,omitempty"`
    Username string `json:"username" bson:"username"`
    Property Property `json:"property" bson:"property"`
    Retweeted int `json:"retweeted" bson:"retweeted"`
    Content string `json:"content" bson:"content"`
    Timestamp int64 `json:"timestamp" bson:"timestamp"`
    ChildType string `json:"childType,omitempty" bson:"childType,omitempty"`
    ParentID objectid.ObjectID `json:"parent,omitempty" bson:"parent,omitempty"`
    MediaIDs []objectid.ObjectID `json:"media,omitempty" bson:"media,omitempty"`
}

type Property struct {
    Likes int `json:"likes" bson:"likes"`
}
/*
func (it Item) MarshalJSON() ([]byte, error) {
    iod := objectid.FromHex(
    m := map[string]interface{}{
        "_id": FromHex(it.
    }
}
*/
