package handlers

import (
    "fmt"
    "net/http"
	"context"
	"time"
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
	"github.com/techswarn/playserver/utils"
)


var cs *kubernetes.Clientset
func init() {
    cs, _ = utils.GetKubehandle()
}


type Deploy struct {
	Name string `json:"name"`
	CreatedAt time.Time `json:"createdate"`
}

func Health(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		Message string
	}
	res := Response{ Message: "API is live"}
	log.Println("Serving:", r.URL.Path, "from", r.Host)
	//Set Content-Type header so that clients will know how to read response
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	resJson, _ := json.Marshal(res)
	w.Write(resJson)
}

func CreatPodHandler(w http.ResponseWriter, r *http.Request){
	fmt.Println(r.Method)
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// If this handler is called create a pod and send a JSON with necessary pod details


	deploy := Deploy{}

	err := json.NewDecoder(r.Body).Decode(&deploy)
	if err != nil{
		panic(err)
	}

	deploy.CreatedAt = time.Now().Local()
    //create a namespace named "foo" and delete it when main exits
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	id := uuid.New()

	ns := "app-" + id.String()
	fmt.Println(ns)
	namespace := createNamespace(ctx, cs, ns)
    log.Printf("Create namespace response: %#v \n", namespace)
	//create an nginx deployment named "hello-world" in the nsFoo namespace
	deployment := deployNginx(ctx, cs, namespace, deploy.Name)
	log.Printf("Is deployed %#v", deployment)

	if(deployment.Status) {
		w.Header().Set("Content-Type","application/json")
		w.WriteHeader(http.StatusCreated)
	    deployRes, err := json.Marshal(deployment)
		if err != nil{
			panic(err)
		}
		w.Write(deployRes)
	} else {
		w.Header().Set("Content-Type","application/json")
		w.WriteHeader(http.StatusInternalServerError)
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

type Replica struct {
	Name string
	Status bool
}

func waitForReadyReplicas(ctx context.Context, clientSet *kubernetes.Clientset, deployment *appv1.Deployment) *Replica {

	fmt.Printf("Waiting for ready replicas in deployment %q\n", deployment.Name)
	for {
		expectedReplicas := *deployment.Spec.Replicas
		readyReplicas := getReadyReplicasForDeployment(ctx, clientSet, deployment)
		if readyReplicas == expectedReplicas {
			fmt.Printf("replicas are ready!\n\n")
			return &Replica{
				Name: deployment.Name,
				Status: true,
			}
			break
		}

		fmt.Printf("replicas are not ready yet. %d/%d\n", readyReplicas, expectedReplicas)
		time.Sleep(1 * time.Second)
	}

	return &Replica{
		Name: "",
		Status: false,
	}
}

func getReadyReplicasForDeployment(ctx context.Context, clientSet *kubernetes.Clientset, deployment *appv1.Deployment) int32 {
	dep, err := clientSet.AppsV1().Deployments(deployment.Namespace).Get(ctx, deployment.Name, metav1.GetOptions{})
	panicIfError(err)

	return dep.Status.ReadyReplicas
}

func deployNginx(ctx context.Context, clientSet *kubernetes.Clientset, ns *corev1.Namespace, name string) *Replica {
	deployment := createNginxDeployment(ctx, clientSet, ns, name)
	s := waitForReadyReplicas(ctx, clientSet, deployment)
	return s
}

func panicIfError(err error) {
	if err != nil {
		panic(err.Error())
	}
}