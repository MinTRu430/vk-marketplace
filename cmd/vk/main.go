package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"VK/internal/ad"
	"VK/internal/session"
	"VK/internal/user"
	"VK/internal/utils/dbutils"
)

var (
	dsn = "postgres://user:password@localhost:5432/vkontakte" // вынести в отдельную функцию + настройки закинуть в env
)

func main() {
	ctx := context.Background()
	fmt.Println("start vk")

	db, err := dbutils.NewPostresPool(ctx, dsn)
	if err != nil {
		log.Fatalln("DB connection failed:", err)
	}
	defer db.Close()

	userDB := user.NewUserRepository(db)
	sessionDB := session.NewSessionsDB(db)
	adRepo := ad.NewAdRepo(db)

	u := &user.UserHandler{
		Sessions: sessionDB,
		UserDB:   userDB,
	}

	ad := &ad.AdHandler{
		Ads:      adRepo,
		Sessions: sessionDB,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/user/reg", u.Registration)
	mux.HandleFunc("/user/login", u.Login)
	mux.HandleFunc("/user/logout", u.Logout)

	mux.HandleFunc("/ads/create", ad.CreateAd)
	mux.HandleFunc("/ads/list", ad.ListAds)

	listenAddr := ":8080"
	log.Printf("starting listening server at %s", listenAddr)
	http.ListenAndServe(listenAddr, mux)

}
