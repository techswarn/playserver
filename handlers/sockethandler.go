package handlers

import (
	"log"
	"net/http"
	"github.com/gorilla/websocket"
	"encoding/json"
	"strings"
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
	type Input struct {
		Kind string
		Message string
	}
	s := []string{}
    for {
		input := &Input{}
        messageType, p, err := conn.ReadMessage()		
		_ = messageType
		myString := string(p)

		err = json.Unmarshal([]byte(myString), &input)
		log.Printf("INPUT struct %#v \n", input)
		if err != nil {
			log.Println("Error parsing JSON:", err)
			return
		}
		if input.Message != "\r" && input.Message != "\u0003" &&  input.Message != "\x7f"{
			s = append(s, input.Message)
		}
		if len(s) > 0 && input.Message == "\x7f" {
			log.Println("REMOVE ELEMENT")
			s = s[:len(s)-1]
			if err := conn.WriteMessage(messageType, []byte("\b \b")); err != nil {
				log.Println(err)
				return
			}
			
	    }

		cmd := strings.Join(s, "")
		log.Println("length:", len(cmd))
		trimmedStr := strings.TrimLeft(cmd, " ")


		log.Println("key=========:",input.Message)
		log.Println("length:", len(s))
		log.Printf("COMMAND %s \n", s)
		switch input.Message {
		case "\x03":
			log.Println("Send SIGINT")
			
			outstr, _, _ := ExeCmd(trimmedStr)
			log.Printf("outstr %s \n", outstr)
			s = []string{}
			if err != nil {
				log.Printf("CLOSED: %s \n", err)
			  //  deleteNamespace(ctx, cs, nsFoo)
				return
			}
			if err := conn.WriteMessage(messageType, []byte(outstr)); err != nil {
				log.Println(err)
				return
			}
		case "\r":
			log.Println("Send message to console")
			log.Println("Command=========:",trimmedStr)
			log.Println("length:", len(trimmedStr))
//			if len(trimmedStr) == 0 {
				if err := conn.WriteMessage(messageType, []byte("\r\n")); err != nil {
					log.Println(err)
					return
				}
//			}
			outstr, _, _ := ExeCmd(trimmedStr)
			log.Printf("outstr ============= %s \n", outstr)

			if err != nil {
				log.Printf("CLOSED: %s \n", err)
			  //  deleteNamespace(ctx, cs, nsFoo)
				return
			}
			if err := conn.WriteMessage(messageType, []byte(outstr)); err != nil {
				log.Println(err)
				return
			}

			if err := conn.WriteMessage(messageType, []byte("root@ubuntu:~$ ")); err != nil {
				log.Println(err)
				return
			}
			s = []string{}
		// case "\x7f":

		// 	// if len(trimmedStr) > 0 {
		// 	// 	log.Println("REMOVE ELEMENT")
		// 	// 	trimmedStr = trimmedStr[:len(trimmedStr)-1]
		// 	// 	if err := conn.WriteMessage(messageType, []byte("\b \b")); err != nil {
		// 	// 		log.Println(err)
		// 	// 		return
		// 	// 	}
		// 	// }
		// }
		}
    }
}