package main

import (
    "fmt"
    "net/http"
	//"context"
	"time"
	"os"
	//"strconv"
	//"log"
	"github.com/techswarn/playserver/handlers"
	"github.com/techswarn/playserver/database"
	"github.com/techswarn/playserver/utils"
	 "github.com/rs/cors"
)

func main() {
	PORT := ":8001"
	arguments := os.Args
	if len(arguments) != 1 {
		PORT = ":" + arguments[1]
	}
	fmt.Println("Using port number: ", PORT)
	database.InitDatabase(utils.GetValue("DB_NAME"))
	mux := http.NewServeMux()

	mux.Handle("/api/v1/health", http.HandlerFunc(handlers.Health))
	mux.Handle("/api/v1/deploy", http.HandlerFunc(handlers.CreatPodHandler))
//	mux.Handle("/api/v1/exec", http.HandlerFunc(handlers.ExeCmd))
	mux.Handle("/ws", http.HandlerFunc(handlers.SocketHandler))
    // http.HandleFunc("/", handler)
	// http.HandleFunc("/api", creatPodHandler)
    handler := cors.Default().Handler(mux)
	s := &http.Server{
		Addr:         PORT,
		Handler:      handler,
		IdleTimeout:  100 * time.Second,
		ReadTimeout:  100 * time.Second,
		WriteTimeout: 100 * time.Second,
	}
	err := s.ListenAndServe()
	if err != nil {
		fmt.Println(err)
		return
	}
}

