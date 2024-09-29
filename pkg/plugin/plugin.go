package plugin

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

func RunPlugin(configFlags *genericclioptions.ConfigFlags) error {
	config, err := configFlags.ToRESTConfig()
	if err != nil {
		return fmt.Errorf("failed to read kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create clientset: %w", err)
	}

	if err := CrashedPods(clientset); err != nil {
		return fmt.Errorf("failed to get crashed pods: %w", err)
	}

	return nil
}

func CrashedPods(clientset *kubernetes.Clientset) error {
	pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "POD NAME\tNAMESPACE\tSTATUS\tLOGS")

	for _, pod := range pods.Items {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.State.Waiting != nil && containerStatus.State.Waiting.Reason == "CrashLoopBackOff" {
				log := GetLogsOfPod(clientset, pod)
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", pod.Name, pod.GetNamespace(), containerStatus.State.Waiting.Reason, log)
			}
		}
	}

	return nil
}

func GetLogsOfPod(clientset *kubernetes.Clientset, pod v1.Pod) string {
	req := clientset.CoreV1().Pods(pod.GetNamespace()).GetLogs(pod.Name, &v1.PodLogOptions{})
	podLogs, err := req.Stream()
	if err != nil {
		fmt.Println("Error in opening stream")
	}

	defer podLogs.Close()

	buf, err := io.ReadAll(podLogs)
	if err != nil {
		fmt.Println("Error in reading stream")
		return "Error in reading stream"
	}

	log := string(buf)
	return log
}
