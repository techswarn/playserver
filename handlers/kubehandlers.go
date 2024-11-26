package handlers

import (
    "fmt"
    "net/http"
    "github.com/gorilla/websocket"
	"context"
	"time"
	"os"
	"encoding/json"
	"github.com/google/uuid"
	//"strconv"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
//	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "github.com/Azure/go-autorest/autorest/to"
	"log"
)

func Createpod(){
	log.Println("create pod called")
}