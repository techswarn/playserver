package handlers

import (
    "fmt"
    "net/http"
	"context"
	"time"
	"encoding/json"
	"github.com/google/uuid"
	"bytes"
	"strconv"
    appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
//	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	//restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
    "github.com/Azure/go-autorest/autorest/to"
	"log"
	"github.com/techswarn/playserver/utils"
	"k8s.io/client-go/tools/clientcmd"
	"math/rand"
	"strings"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/techswarn/playserver/models"
	"github.com/techswarn/playserver/database"
)

//Create clients
var cs *kubernetes.Clientset
var rs *redis.Client

func init() {
    cs, _ = utils.GetKubehandle()
	rs  = utils.Client()
}

type JobInfo struct {
	JobId string `json:"jobid"`
	DeployID string  `json:"deployid"`
}

//HEALTH CHECK HANDLER
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

//ADD DEPLOY DETAILS TO REDIS QUEUE

func CreatPodHandler(w http.ResponseWriter, r *http.Request){
	fmt.Println(r.Method)
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// If this handler is called create a pod and send a JSON with necessary pod details


	deployRequest := models.DeployRequest{}

	err := json.NewDecoder(r.Body).Decode(&deployRequest)
	if err != nil{
		panic(err)
	}
	log.Printf("Deploy request:--- %#v \n", deployRequest)

    //create a namespace named "foo" and delete it when main exits
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	ns := "app-" + uuid.New().String()
	fmt.Println(ns)

	//deployId := uuid.New() 
	//deploy.CreatedAt = time.Now().Local()
	namespace := createNamespace(ctx, cs, ns)
	log.Printf("Creating namespace %#v \n", namespace)
		// validate the request
    errors := deployRequest.ValidateStruct()
	log.Printf("Validation error %#v \n", errors)
	//if validation is failed, return the validation errors
	if errors != nil {
		w.Header().Set("Content-Type","application/json")
		w.WriteHeader(http.StatusBadRequest)
	    res := &models.Response[[]*models.ErrorResponse]{
			Success: false,
			Message: "validation failed",
			Data: errors,
		} 
		insertdeploy, err := json.Marshal(res)
		if err != nil{
			panic(err)
		}
		w.Write(insertdeploy)		
	}
	var deploy models.Deploy
	deploy = models.Deploy{
		Id : uuid.New().String(),
		Name : deployRequest.Name,
		Image : deployRequest.Image,
		Namespace : ns,
		CreatedAt : time.Now(),
	}
	log.Printf("Deployment details %#v", deploy)
	if result := database.DB.Create(&deploy); result.Error != nil {
		fmt.Printf("DB write error: %s", &result.Error)
		panic(&result.Error)
	}
	jobID := strconv.Itoa(rand.Intn(1000) + 1)
	jobInfo := JobInfo{JobId: jobID, DeployID: deploy.Id}
	job, err := json.Marshal(jobInfo)
	if err != nil {
		log.Fatal(err)
	}

	err = rs.LPush(context.Background(), "jobs", job).Err()
	if err != nil {
		log.Fatal("lpush issue", err)
	}


	//As soon as the details is incerted to db poluate the value in redis queue


	// result := database.DB.Create(&deploy)
	// log.Printf("RESULT %#v", result)
	res := &models.Response[*models.Deploy]{
		Success: true,
		Message: "Queued deployment",
		Data: &deploy,
	} 
	//Return the deployment data stored
	 w.Header().Set("Content-Type","application/json")
	 w.Header().Add("jobid", jobID)
	 w.WriteHeader(http.StatusCreated)
	 deployRes, err := json.Marshal(res)
	 if err != nil{
		panic(err)
	 }
	w.Write(deployRes)


//	log.Printf("Deployment details %#v", deploy)
	//create an nginx deployment named "hello-world" in the nsFoo namespace
//	deployment, s := CreateDeploy(ctx, cs, namespace, deployRequest.Name, deployRequest.Image)
//	log.Printf("Is deployed %#v", s)
//	log.Printf("Deployment details %#v", deployment.ObjectMeta)

	//Get deployment details and then update MYSQL DB
	//deployID := string(deployment.ObjectMeta.UID)
	//CreateAt := time.Now()

	// log.Printf("Update details to db %#v", deploy)
	// if(s.Status) {
	// 	w.Header().Set("Content-Type","application/json")
	// 	w.WriteHeader(http.StatusCreated)
	//     deployRes, err := json.Marshal(deployment)
	// 	if err != nil{
	// 		panic(err)
	// 	}
	// 	w.Write(deployRes)
	// } else {
	// 	w.Header().Set("Content-Type","application/json")
	// 	w.WriteHeader(http.StatusInternalServerError)
	// }
}

//Create Namespace:
func createNamespace(ctx context.Context, clientSet *kubernetes.Clientset, name string) *corev1.Namespace {

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
func createNginxDeployment(ctx context.Context, clientSet *kubernetes.Clientset, ns *corev1.Namespace, name string, image string) *appv1.Deployment {
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
							Image: image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
								},						
							},
							Command: []string{"/bin/sh", "-c", "sleep 50000"},
						},
					},
				},
			},
		},
	}
	deployment, err := clientSet.AppsV1().Deployments(ns.Name).Create(ctx, deployment, metav1.CreateOptions{})
	//log.Printf(" deployment: %#v", deployment)
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

func CreateDeploy(ctx context.Context, clientSet *kubernetes.Clientset, ns *corev1.Namespace, name string, image string) (*appv1.Deployment, *Replica) {
	deployment := createNginxDeployment(ctx, clientSet, ns, name, image)
	s := waitForReadyReplicas(ctx, clientSet, deployment)
	return deployment, s
}

func panicIfError(err error) {
	if err != nil {
		panic(err.Error())
	}
}


//EXECUTE COMMAND HANDLER
func ExeCmd(cmd string) (string, int, error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pods := getpods(ctx, cs, "app-dc236ebb-9a03-40d5-99c3-828464f92208")
	fmt.Printf("string: %s \n", pods.Items[0].Name)
	outstr, errstr, err := ExecuteRemoteCommand("app-dc236ebb-9a03-40d5-99c3-828464f92208", pods.Items[0].Name, cmd)
	return outstr, errstr, err
}

//GET PODS
func getpods(ctx context.Context, clientSet *kubernetes.Clientset, ns string) *corev1.PodList {
	pods, err := clientSet.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
	panicIfError(err)
	//fmt.Printf("=====PODS======%#v \n", pods)
	// for _, pod := range pods.Items {
	// 	fmt.Printf("Pod name: %v \n", pod)
	// }
	return pods
}
//pod *corev1.Pod, 
type LogStreamer struct{
    b bytes.Buffer
}

func (l *LogStreamer) String() string {
    return l.b.String()
}

func (l *LogStreamer) Write(p []byte) (n int, err error) {
    a := string(p)
    l.b.WriteString(a)
    log.Println(a)
    return len(p), nil
}

func ExecuteRemoteCommand( ns string, pod string, command string) (string, int, error) {
  //  fmt.Println(ns)
	//fmt.Println(pod)
	fmt.Printf("Command: %s \n", command)

	
	kubeCfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	restCfg, err := kubeCfg.ClientConfig()
	if err != nil {
		return "", 1, err
	}
	coreClient, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return "", 1, err
	}

//	buf := &bytes.Buffer{}
//	errBuf := &bytes.Buffer{}
	request := coreClient.CoreV1().RESTClient().
		Post().
		Namespace(ns).
		Resource("pods").
		Name(pod).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command: []string{"/bin/sh", "-c", command},
			Stdin:     true,
			Stdout:    true,
			Stderr:    false,
			TTY:       true,
		}, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(restCfg, "POST", request.URL())
	if err != nil {
		log.Printf("Error during NewSPDYExecutor %s \n", err)
	}
    var streamErr error
    l := &LogStreamer{}
	streamErr = exec.Stream(remotecommand.StreamOptions{
        Stdin:  os.Stdin,
        Stdout: l,
        Stderr: nil,
        Tty:    true,
    })
	log.Printf("OUTPUT STREAM ================ %s \n", l)

    if streamErr != nil {
        if strings.Contains(streamErr.Error(), "command terminated with exit code") {
            return l.String(), 1, nil
        } else {
            return "", 0, fmt.Errorf("could not stream results: %w", streamErr)
        }
    }

	return l.String(), 0, nil
}
