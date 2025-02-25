package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	sessionRepo "github.com/lightlink/auth-service/internal/session/repository/redis"
	userRepo "github.com/lightlink/auth-service/internal/user/repository/grpc"
	proto "github.com/lightlink/auth-service/protogen/user"

	sessionUsecase "github.com/lightlink/auth-service/internal/session/usecase"

	sessionDelivery "github.com/lightlink/auth-service/internal/session/delivery/http"
)

func main() {
	client, _ := grpc.Dial(
		fmt.Sprintf("%s:%s", os.Getenv("USER_SERVICE_HOST"), os.Getenv("USER_SERVICE_PORT")),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	userServiceClient := proto.NewUserServiceClient(client)
	userRepository := userRepo.NewUserGrpcRepository(&userServiceClient)

	redisURL := fmt.Sprintf("redis://user:@%s:%s/%s",
		os.Getenv("REDIS_HOST"),
		os.Getenv("REDIS_PORT"),
		os.Getenv("REDIS_DATABASE"),
	)
	redisConn, err := redis.DialURL(redisURL)
	if err != nil {
		panic(err)
	}

	sessionRepository := sessionRepo.NewSessionRedisRepository(redisConn)

	sessionUsecase := sessionUsecase.NewSessionUsecase(
		sessionRepository,
		userRepository,
	)

	sessionHandler := sessionDelivery.NewSessionHandler(sessionUsecase)

	router := mux.NewRouter()

	router.HandleFunc("/api/signup", sessionHandler.Signup).Methods("POST")
	router.HandleFunc("/api/login", sessionHandler.Login).Methods("POST")
	router.HandleFunc("/api/logout", sessionHandler.Logout).Methods("POST")
	router.HandleFunc("/api/refresh", sessionHandler.Refresh).Methods("GET")
	router.HandleFunc("/api/check", sessionHandler.Check).Methods("GET")

	log.Println("starting server at http://127.0.0.1:8082")
	log.Fatal(http.ListenAndServe(":8082", router))
}
