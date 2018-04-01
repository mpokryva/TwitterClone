package user

type User struct {
    Username string `json:"username,omitempty" bson:"username"`
    Email string `json:"email,omitempty" bson:"email"`
    Password string `json:"password,omitempty" bson:"password"`
    Followers int `json:"followers" bson:"followers"`
    Following int `json:"following" bson:"following"`
    Verified bool `json:"verified,omitempty" bson:"verified"`
    Key string `json:"key,omitempty" bson:"key"`
}

