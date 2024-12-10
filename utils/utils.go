package utils

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
    metrics "k8s.io/metrics/pkg/client/clientset/versioned"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"

)

//REDIS IMPORTS
var client *redis.Client

func Client() *redis.Client {
	fmt.Println(GetValue("REDIS_URL"))
	url := GetValue("REDIS_URL")
    opts, err := redis.ParseURL(url)
    if err != nil {
        panic(err)
    }
    return redis.NewClient(opts)
}

func GetKubehandle() (*kubernetes.Clientset, *metrics.Clientset) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		fmt.Println(home)
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		fmt.Println(home)
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	mc, err := metrics.NewForConfig(config)
    if err != nil {
        panic(err.Error())
    }
	return clientset, mc
}

func CheckError(err error) {
	if err != nil {
		log.Fatalf("Get: %v", err)
	}
}


func GetValue(key string) string {
	fmt.Println(os.Getenv("GO_ENV"))
	env := os.Getenv("GO_ENV")
    // load the .env file
	fmt.Printf("The env value is %s \n", env)

	if os.Getenv("GO_ENV") != "PRODUCTION" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatalf("Error loading .env file!!\n")
		}
	}

    // return the value based on a given key
	return os.Getenv(key)
}