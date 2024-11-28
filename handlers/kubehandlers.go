package handlers

import (
    "fmt"
    "net/http"
	"context"
	"time"
	"encoding/json"
	"github.com/google/uuid"
	"bytes"
	//"strconv"
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

	"strings"
	"os"

)


var cs *kubernetes.Clientset
func init() {
    cs, _ = utils.GetKubehandle()
}


type Deploy struct {
	Name string `json:"name"`
	CreatedAt time.Time `json:"createdate"`
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

//CREATE DEPLOYMENT HANDLER
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
	//create an nginx deployment named "hello-world" in the nsFoo namespace
	deployment := deployNginx(ctx, cs, namespace, deploy.Name)
	//log.Printf("Is deployed %#v", deployment)

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
							Image: "ubuntu:latest",
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
	//fmt.Println(command)

	
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
    if streamErr != nil {
        if strings.Contains(streamErr.Error(), "command terminated with exit code") {
            return l.String(), 1, nil
        } else {
            return "", 0, fmt.Errorf("could not stream results: %w", streamErr)
        }
    }

	return l.String(), 0, nil
}
