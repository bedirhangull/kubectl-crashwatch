package plugin

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

func RunDashboard(configFlags *genericclioptions.ConfigFlags) error {

	config, err := configFlags.ToRESTConfig()
	if err != nil {
		return fmt.Errorf("failed to read kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create clientset: %w", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		crashedPodsHandler(w, r, clientset)
	})

	http.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) {
		podLogsHandler(w, r, clientset)
	})

	fmt.Println("Starting server at http://localhost:8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	return err
}

func crashedPodsHandler(w http.ResponseWriter, r *http.Request, clientset *kubernetes.Clientset) {
	podsInfo, err := GetCrashedPodsInfo(clientset)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get pods: %v", err), http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.New("dashboard").Parse(dashboardTemplate))
	err = tmpl.Execute(w, podsInfo)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
		return
	}
}

func podLogsHandler(w http.ResponseWriter, r *http.Request, clientset *kubernetes.Clientset) {
	namespace := r.URL.Query().Get("namespace")
	podName := r.URL.Query().Get("pod")

	if namespace == "" || podName == "" {
		http.Error(w, "Missing namespace or pod parameter", http.StatusBadRequest)
		return
	}

	pod, err := clientset.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get pod: %v", err), http.StatusInternalServerError)
		return
	}

	logContent := GetLogsOfPod(clientset, *pod)

	data := struct {
		Namespace string
		PodName   string
		Log       string
	}{
		Namespace: namespace,
		PodName:   podName,
		Log:       logContent,
	}

	tmpl := template.Must(template.New("logs").Parse(logsTemplate))
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
		return
	}
}

func GetCrashedPodsInfo(clientset *kubernetes.Clientset) ([]v1.Pod, error) {
	pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	var crashedPods []v1.Pod
	for _, pod := range pods.Items {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.State.Waiting != nil && containerStatus.State.Waiting.Reason == "CrashLoopBackOff" {
				crashedPods = append(crashedPods, pod)
				break
			}
		}
	}

	return crashedPods, nil
}

const dashboardTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>CrashLoopBackOff Pods Dashboard</title>
    <style>
        body { font-family: Arial, sans-serif; }
        table { width: 100%; border-collapse: collapse; margin-bottom: 20px; }
        th, td { padding: 8px 12px; border: 1px solid #ccc; }
        th { background-color: #f4f4f4; }
        a { text-decoration: none; color: #007BFF; }
    </style>
</head>
<body>
    <h1>CrashLoopBackOff Pods Dashboard</h1>
    <table>
        <thead>
            <tr>
                <th>Namespace</th>
                <th>Pod Name</th>
                <th>Status</th>
            </tr>
        </thead>
        <tbody>
            {{range .}}
            <tr>
                <td>{{.Namespace}}</td>
                <td><a href="/logs?namespace={{.Namespace}}&pod={{.Name}}">{{.Name}}</a></td>
                <td>{{range .Status.ContainerStatuses}}{{if eq .State.Waiting.Reason "CrashLoopBackOff"}}{{.State.Waiting.Reason}}{{end}}{{end}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>
</body>
</html>
`

const logsTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Logs for {{.PodName}}</title>
    <style>
        body { font-family: Arial, sans-serif; }
        pre { background-color: #f8f8f8; padding: 10px; }
        a { text-decoration: none; color: #007BFF; }
    </style>
</head>
<body>
    <h1>Logs for {{.Namespace}}/{{.PodName}}</h1>
    <a href="/">Back to Dashboard</a>
    <pre>{{.Log}}</pre>
</body>
</html>
`
