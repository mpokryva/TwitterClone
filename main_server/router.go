package main

import (
    "net/http"
    "github.com/gorilla/mux"
    "TwitterClone/additem"
)

type Route struct {
    Name string
    Method string
    Path string
    HandlerFunc http.HandlerFunc
}

var routes = []Route{
    Route {
        "additem",
        "POST",
        "/additem",
        additem.AddItemHandler},
    Route {
        "adduser",
        "POST",
        "/adduser",
        adduser.AddUserHandler},
    Route {
        "login",
        "POST",
        "/login",
        login.LoginHandler},
    Route {
        "logout",
        "POST",
        "/logout",
        logout.LogoutHandler},
    Route {
        "verify",
        "POST",
        "/verify",
        verify.VerifyHandler},
    Route {
        "GetItem",
        "GET",
        "/item/{id}",
        item.GetItemHandler},
    Route {
        "LikeItem",
        "POST",
        "/item/{id}/like",
        item.LikeItemHandler},
    Route {
        "DeleteItem",
        "DELETE",
        "/item/{id}",
        item.DeleteItemHandler},
    Route {
        "search",
        "POST",
        "/search",
        search.SearchHandler},
    Route {
        "GetUser",
        "GET",
        "/user/{username}",
        user.GetUserHandler},
    Route {
        "GetUserFollowers",
        "GET",
        "/user/{username}/followers",
        user.GetFollowersHandler},
    Route {
        "GetUserFollowing",
        "GET",
        "/user/{username}/following",
        user.GetFollowingHandler},
    Route {
        "Follow",
        "POST",
        "/follow",
        follow.FollowHandler},
    Route {
        "AddMedia",
        "POST",
        "/addmedia",
        media.AddMediaHandler},
    Route {
        "GetMedia",
        "GET",
        "/media/{id}",
        media.GetMediaHandler},

}
func NewRouter() *mux.Router {
    router := mux.NewRouter()
    for _, route := range routes {
        router.
        Methods(route.Method).
        Path(route.Path).
        Name(route.Name).
        HandlerFunc(route.HandlerFunc)
    }
    return router
}
