package item

import (
    "github.com/mongodb/mongo-go-driver/bson/objectid"
    "encoding/json"
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

type internalItem struct {
    ID string `json:"id" bson:"_id,omitempty"`
    Username string `json:"username" bson:"username"`
    Property Property `json:"property" bson:"property"`
    Retweeted int `json:"retweeted" bson:"retweeted"`
    Content string `json:"content" bson:"content"`
    Timestamp int64 `json:"timestamp" bson:"timestamp"`
    ChildType string `json:"childType,omitempty" bson:"childType,omitempty"`
    ParentID string `json:"parent,omitempty" bson:"parent,omitempty"`
    MediaIDs []string `json:"media,omitempty" bson:"media,omitempty"`

}

type Property struct {
    Likes int `json:"likes" bson:"likes"`
}

func (it Item) MarshalJSON() ([]byte, error) {
    var inIt internalItem
    var nilObjectID objectid.ObjectID
    oid := it.ID.Hex()
    inIt.ID = oid
    if it.ParentID != nilObjectID {
        inIt.ParentID = it.ParentID.Hex()
    }
    for _, mID := range it.MediaIDs {
        mOid := mID.Hex()
        inIt.MediaIDs = append(inIt.MediaIDs, mOid)
    }
    inIt.Username = it.Username
    inIt.Property = it.Property
    inIt.Retweeted = it.Retweeted
    inIt.Content = it.Content
    inIt.Timestamp = it.Timestamp
    inIt.ChildType = it.ChildType
    return json.Marshal(inIt)
}
