package user

type User struct {
    Username string `json:"username,omitempty" bson:"username"`
    Email string `json:"email,omitempty" bson:"email"`
    Password string `json:"password,omitempty" bson:"password"`
    Followers []string `json:"followers" bson:"followers"`
    Following []string `json:"following" bson:"following"`
    FollowerCount int `json:"followerCount" bson:"followerCount"`
    FollowingCount int `json:"followingCount" bson:"followingCount"`
    Verified bool `json:"verified,omitempty" bson:"verified"`
    Key string `json:"key,omitempty" bson:"key"`
}

