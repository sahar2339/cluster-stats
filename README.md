# Kubernetes Cluster Scanner
This repository contains a script that scans a Kubernetes cluster and generates a CSV file that provides insights into the resource utilization and distribution of containers across namespaces, as well as average CPU and memory allocation across the cluster's nodes.

# Usage
Clone the repository to your local machine:

```
git clone https://github.com/your-username/kubernetes-cluster-scanner.git
```
Set the KUBECONFIG environment variable to point to your Kubernetes configuration file:
```
oc login
``` 
Run the script:
```
./dist/main
```
After the script completes, a CSV file named cluster_containers.csv will be generated in the same directory. This file provides information on the distribution of container instances across different sizes and namespaces, as well as average CPU and memory allocation on cluster nodes.

# Purpose and Flow
The purpose of this script is to provide insights into the resource utilization and distribution of containers in a Kubernetes cluster, as well as the average CPU and memory allocation across the cluster's nodes. This information can be useful for optimizing resource allocation, identifying potential bottlenecks, and gaining an understanding of resource usage patterns within the cluster.

The script follows the following flow:

1. Establishes a connection to the Kubernetes cluster using the provided KUBECONFIG file.
2. Retrieves the list of pods in the cluster.
3. Iterates over each pod and container, extracting the resource limits (CPU and memory) for each container.
4. Groups the containers by their resource limits into predefined instance sizes (e.g., small, medium, large).
5. Tracks the distribution of container instances across namespaces and instance sizes.
6. Retrieves the list of nodes in the cluster.
7. Calculates the average CPU and memory allocation across all nodes, taking into account the allocated resources and the pods running on each node.
8. Writes the distribution of container instances and node information to a CSV file named cluster_containers.csv.
The resulting CSV file provides a summary of container distribution and node utilization, helping cluster administrators gain insights into the resource allocation and utilization patterns within their Kubernetes cluster.
