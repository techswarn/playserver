package handlers

import (
	"log"
	"net/http"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

func SocketHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("socketHandler")
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
        return
    }
    defer conn.Close()

    //If a request comes to websocket create a namespace and pod.
    // ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
    //create a namespace named "foo" and delete it when main exits
	// nsFoo := createNamespace(ctx, cs, "foo")
    // log.Printf("Create namespace response: %#v \n", nsFoo)
	// // create an nginx deployment named "hello-world" in the nsFoo namespace
	// isDeployed := deployNginx(ctx, cs, nsFoo, "hello-world")
	// log.Println("Is deployed ", isDeployed)

    for {

        messageType, p, err := conn.ReadMessage()
		myString := string(p)
		outstr, _, _ := ExeCmd(myString)
		
	    log.Println(myString)

        if err != nil {
            log.Printf("CLOSED: %s \n", err)
          //  deleteNamespace(ctx, cs, nsFoo)
            return
        }
        if err := conn.WriteMessage(messageType, []byte(outstr)); err != nil {
            log.Println(err)
            return
        }
    }
}