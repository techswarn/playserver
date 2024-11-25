package main

import (
    "fmt"
    "net/http"
    "github.com/gorilla/websocket"
	"context"
	"time"
	"os"
	//"strconv"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
//	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "github.com/Azure/go-autorest/autorest/to"
)

var cs *kubernetes.Clientset
func init() {
    cs, _ = getKubehandle()
}


//Add a Code block a setup socket connect to Kubernetes pod.

//podurl := "wss://127.0.0.1:8080/api/v1/namespaces/default/pods/backend-7b8bb8b977-hwpqj/exec?command=sh&stdin=true&stdout=true&tty=true";

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
        return true
    },
}



func creatPodHandler(w http.ResponseWriter, r *http.Request){
	// If this handler is called create a pod and send a JSON with necessary pod details
	fmt.Fprintf(w, "Serving: %s\n", r.URL.Path)
	fmt.Printf("Served: %s\n", r.Host)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("handler")
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer conn.Close()

    //If a request comes to websocket create a namespace and pod.
    // ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
    //create a namespace named "foo" and delete it when main exits
	// nsFoo := createNamespace(ctx, cs, "foo")
    // fmt.Printf("Create namespace response: %#v \n", nsFoo)
	// // create an nginx deployment named "hello-world" in the nsFoo namespace
	// isDeployed := deployNginx(ctx, cs, nsFoo, "hello-world")
	// fmt.Println("Is deployed ", isDeployed)

    for {
        messageType, p, err := conn.ReadMessage()
		myString := string(p)
	    fmt.Println(myString)

        if err != nil {
            fmt.Printf("CLOSED: %s \n", err)
          //  deleteNamespace(ctx, cs, nsFoo)
            return
        }
        if err := conn.WriteMessage(messageType, p); err != nil {
            fmt.Println(err)
            return
        }
    }
}

func main() {
	PORT := ":8001"
	arguments := os.Args
	if len(arguments) != 1 {
		PORT = ":" + arguments[1]
	}
	fmt.Println("Using port number: ", PORT)
	
	fmt.Println("Using port number: ", PORT)
    http.HandleFunc("/", handler)
	http.HandleFunc("/api", creatPodHandler)

	err := http.ListenAndServe(PORT, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
}

//Create Namespace:
func createNamespace(ctx context.Context, clientSet *kubernetes.Clientset, name string) *corev1.Namespace {
	fmt.Printf("Creating namespace %q.\n\n", name)
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	ns, err := clientSet.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	panicIfError(err)
	return ns
}

//DELETE NAMESPACE
func deleteNamespace(ctx context.Context, clientSet *kubernetes.Clientset, ns *corev1.Namespace) {
	fmt.Printf("\n\nDeleting namespace %q.\n", ns.Name)
	panicIfError(clientSet.CoreV1().Namespaces().Delete(ctx, ns.Name, metav1.DeleteOptions{}))
}

//CREATE DEPLOYMENT
func createNginxDeployment(ctx context.Context, clientSet *kubernetes.Clientset, ns *corev1.Namespace, name string) *appv1.Deployment {
	var (
		matchLabel = map[string]string{"app": "nginx"}
		objMeta    = metav1.ObjectMeta{
			Name:      name,
			Namespace: ns.Name,
			Labels:    matchLabel,
		}
	)

	deployment := &appv1.Deployment{
		ObjectMeta: objMeta,
		Spec: appv1.DeploymentSpec{
			Replicas: to.Int32Ptr(1),
			Selector: &metav1.LabelSelector{MatchLabels: matchLabel},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: matchLabel,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: "nginxdemos/hello:latest",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}
	deployment, err := clientSet.AppsV1().Deployments(ns.Name).Create(ctx, deployment, metav1.CreateOptions{})
	panicIfError(err)
	return deployment
}

func waitForReadyReplicas(ctx context.Context, clientSet *kubernetes.Clientset, deployment *appv1.Deployment) bool {
	fmt.Printf("Waiting for ready replicas in deployment %q\n", deployment.Name)
	for {
		expectedReplicas := *deployment.Spec.Replicas
		readyReplicas := getReadyReplicasForDeployment(ctx, clientSet, deployment)
		if readyReplicas == expectedReplicas {
			fmt.Printf("replicas are ready!\n\n")
			return true
			break
		}

		fmt.Printf("replicas are not ready yet. %d/%d\n", readyReplicas, expectedReplicas)
		time.Sleep(1 * time.Second)
	}

	return false
}

func getReadyReplicasForDeployment(ctx context.Context, clientSet *kubernetes.Clientset, deployment *appv1.Deployment) int32 {
	dep, err := clientSet.AppsV1().Deployments(deployment.Namespace).Get(ctx, deployment.Name, metav1.GetOptions{})
	panicIfError(err)

	return dep.Status.ReadyReplicas
}

func deployNginx(ctx context.Context, clientSet *kubernetes.Clientset, ns *corev1.Namespace, name string) bool{
	deployment := createNginxDeployment(ctx, clientSet, ns, name)
	s := waitForReadyReplicas(ctx, clientSet, deployment)
	return s
}

func panicIfError(err error) {
	if err != nil {
		panic(err.Error())
	}
}