package main

import (
	"context"
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
	log.Println("start vk")

	db, err := dbutils.NewPostresPool(ctx, dsn)
	if err != nil {
		log.Fatalln("DB connection failed:", err)
	}
	defer db.Close()
	log.Println("database connected")

	userDB := user.NewUserDB(db)
	sessionDB := session.NewSessionsDB(db)
	adDB := ad.NewAdRepo(db)

	u := &user.UserHandler{
		SessionsDB: sessionDB,
		UserDB:     userDB,
	}

	ad := &ad.AdHandler{
		Ads:      adDB,
		Sessions: sessionDB,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/signin", u.Registration)
	mux.HandleFunc("/login", u.Login)

	mux.HandleFunc("/create", ad.CreateAd)
	mux.HandleFunc("/", ad.ListAds)

	log.Println("Starting server at :8080")
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal("Server failed: ", err)
	}

}
