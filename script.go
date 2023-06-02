package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	csvPath = "cluster_containers.csv"
)

type Resources struct {
	CPU    string
	Memory string
}

type NodeInfo struct {
	Number int
	CPU    float64
	Memory float64
}

func disableWarnings() {
	os.Setenv("KUBERNETES_TRUST_CERT", "true")
}

func getInstanceSize(resources Resources) int {
	switch resources {
	case Resources{CPU: "1", Memory: "8Gi"}:
		return 0
	case Resources{CPU: "2", Memory: "16Gi"}:
		return 1
	case Resources{CPU: "3", Memory: "24Gi"}:
		return 2
	case Resources{CPU: "4", Memory: "32Gi"}:
		return 3
	case Resources{CPU: "8", Memory: "64Gi"}:
		return 4
	case Resources{CPU: "16", Memory: "128Gi"}:
		return 5
	case Resources{CPU: "32", Memory: "256Gi"}:
		return 6
	default:
		return 7
	}
}

func writeToFile(instances map[string][]int, nodeInfo NodeInfo) {
	file, err := os.Create(csvPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	file.WriteString("Namespace / Size, small, medium, large, x-large, xx-large, xxx-large, xxxx-large, other\n")
	for namespace, sizes := range instances {
		namespaceLine := namespace + ","
		for _, size := range sizes {
			if size == 0 {
				namespaceLine = namespaceLine + ","
			} else {
				namespaceLine = namespaceLine + strconv.Itoa(size) + ","
			}
		}
		file.WriteString(namespaceLine + "\n")
	}
	file.WriteString("\n")

	file.WriteString("\n\n\n\n")
	file.WriteString(fmt.Sprintf("Number of nodes %d, Average of CPU limits on nodes: %.2f%%, Average of memory limits on nodes: %.2f%%\n", nodeInfo.Number, nodeInfo.CPU, nodeInfo.Memory))
	fmt.Printf("Finish writing to csv file! (%s)\n", csvPath)
}

func getContainerResources(container v1.Container) Resources {
	resources := Resources{}
	if container.Resources.Limits != nil {
		resources.CPU = container.Resources.Limits.Cpu().String()
		resources.Memory = container.Resources.Limits.Memory().String()
	}
	return resources
}

func parseCPUValue(cpuValue string) int {
	cpuValue = strings.TrimSuffix(cpuValue, "m")
	cpu, _ := strconv.Atoi(cpuValue)
	return cpu
}

func parseMemoryValue(memoryValue string) int64 {
	memoryQuantity, err := resource.ParseQuantity(memoryValue)
	if err != nil {
		return 0
	}

	return memoryQuantity.Value()
}

func calculateNodeAvgAllocation(nodes *v1.NodeList, clientset *kubernetes.Clientset) NodeInfo {
	var (
		totalCPUUtilization    float64
		totalMemoryUtilization float64
		totalNodeCPU           int
		totalNodeMemory        int64
	)

	for _, node := range nodes.Items {
		allocatableResources := node.Status.Allocatable

		cpuValue := allocatableResources["cpu"]
		memoryValue := allocatableResources["memory"]

		totalNodeCPU += parseCPUValue(cpuValue.String())
		totalNodeMemory += parseMemoryValue(memoryValue.String())

		pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
			FieldSelector: "spec.nodeName=" + node.Name,
		})
		if err != nil {
			panic(err)
		}

		totalCPULimits := 0
		var totalMemoryLimits int64
		totalMemoryLimits = 0

		for _, pod := range pods.Items {
			for _, container := range pod.Spec.Containers {
				resources := getContainerResources(container)
				if resources.CPU != "" || resources.Memory != "" {
					totalCPULimits += parseCPUValue(resources.CPU)
					totalMemoryLimits += parseMemoryValue(resources.Memory)
				}
			}
		}

		totalCPUUtilization += float64(totalCPULimits)
		totalMemoryUtilization += float64(totalMemoryLimits)
	}

	averageCPUUtilization := (totalCPUUtilization / float64(totalNodeCPU)) * 100
	averageMemoryUtilization := (totalMemoryUtilization / float64(totalNodeMemory)) * 100
	return NodeInfo{
		Number: len(nodes.Items),
		CPU:    averageCPUUtilization,
		Memory: averageMemoryUtilization,
	}
}

func scanCluster() {
	// Use the in-cluster kubeconfig
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	// Create a Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Print("Make sure you are logged in to the cluster!")
		panic(err)
	}

	instances := make(map[string][]int)

	for _, pod := range pods.Items {
		if pod.DeletionTimestamp != nil || strings.Contains(pod.ObjectMeta.Namespace, "openshift") {
			continue
		}
		for _, container := range pod.Spec.Containers {
			resources := getContainerResources(container)
			if resources.CPU != "" || resources.Memory != "" {
				if _, ok := instances[pod.ObjectMeta.Namespace]; !ok {
					instances[pod.ObjectMeta.Namespace] = make([]int, 8)
				}
				size := getInstanceSize(resources)
				instances[pod.ObjectMeta.Namespace][size]++
			}
		}
	}

	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	nodeInfo := calculateNodeAvgAllocation(nodes, clientset)
	writeToFile(instances, nodeInfo)
}

func main() {
	fmt.Print("Starting...")
	scanCluster()
}
