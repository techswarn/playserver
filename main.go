package main

import (
    "fmt"
    "net/http"
    "github.com/gorilla/websocket"
	_"context"
	_ "time"
	"log"
    "github.com/techswarn/playserver/utils"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)


//Add a Code block a setup socket connect to Kubernetes pod.

//podurl := "wss://127.0.0.1:8080/api/v1/namespaces/default/pods/backend-7b8bb8b977-hwpqj/exec?command=sh&stdin=true&stdout=true&tty=true";

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("handler")
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer conn.Close()

    for {
        messageType, p, err := conn.ReadMessage()
		fmt.Println(messageType)
        if err != nil {
            fmt.Println(err)
            return
        }
        if err := conn.WriteMessage(messageType, p); err != nil {
            fmt.Println(err)
            return
        }
    }
}

func main() {
    cs, _ := utils.getKubehandle()
    _ = cs
    http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}