package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
    metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

func getKubehandle() (*kubernetes.Clientset, *metrics.Clientset) {
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
