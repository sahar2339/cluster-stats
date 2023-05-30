import urllib3
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
from openshift_helper_client import OpenshiftClient

CSV_PATH="cluster_containers.csv"

def get_instance_size(resources) -> str:
    resource__obj = (resources["cpu"], resources["memory"])
    match resource__obj:
        case ("1","8Gi"):
            return "Small"
        case ("2","16Gi"):
            return "Medium"
        case ("3","24Gi"):
            return "Large"
        case ("4","32Gi"):
            return "x-Large"
        case ("8","64Gi"):
            return "xx-Large"
        case ("16","128Gi"):
            return "xxx-Large"
        case ("32","256Gi"):
            return "xxxx-Large"
        case _ :
            return "other"
        

def write_to_csv_file(instances: dict, node_info: dict) -> None :
    with open(CSV_PATH, "w+") as csv_file:
        csv_file.write("")
    with open(CSV_PATH, "a+") as csv_file:
        for size in instances.keys():
            csv_file.write(f"{size},")
        csv_file.write("\n")
        for _,amount in instances.items():
            csv_file.write(f"{amount},")
        csv_file.write("\n")
        csv_file.write("\n\n\n\n")
        csv_file.write(f"Number of nodes {node_info['number']}, Average of CPU limits on nodes: {node_info['cpu']}%, Average of memory limits on nodes: {node_info['memory']}%")
        print(f"Finish writing to csv file! ({CSV_PATH})")


def get_container_resources(container) -> dict:
    if "resources" in container.keys(): 
        if "limits" in container["resources"].keys():
            return container["resources"]["limits"]
    return {}


def parse_cpu_value(cpu_value):
    if cpu_value.endswith("m"):
        # Value is in millicores
        return int(cpu_value[:-1])
    else:
        # Value is in cores, convert to millicores
        return int(float(cpu_value) * 1000)

def parse_memory_value(memory_value):
    if memory_value.endswith("Ki"):
        # Value is in kibibytes, convert to bytes
        return int(memory_value[:-2]) * 1024
    elif memory_value.endswith("Mi"):
        # Value is in mebibytes, convert to bytes
        return int(memory_value[:-2]) * 1024 * 1024
    elif memory_value.endswith("M"):
        # Value is in mebibytes, convert to bytes
        return int(memory_value[:-1]) * 1024 * 1024
    elif memory_value.endswith("Gi"):
        # Value is in gibibytes, convert to bytes
        return int(memory_value[:-2]) * 1024 * 1024 * 1024
    else:
        # Value is already in bytes
        return int(memory_value)



def calculate_node_avg_allocation(nodes, client):
     total_cpu_utilization = 0 
     total_memory_utilization = 0
     total_node_cpu = 0
     total_node_memory = 0

     for node in nodes.items:
        node_name = node.metadata.name
        allocatable_resources = node.status.allocatable
        total_node_cpu += parse_cpu_value(allocatable_resources["cpu"])
        total_node_memory +=  parse_memory_value(allocatable_resources["memory"])

        pods = client.get_resources(kind="Pod",field_selector=f"spec.nodeName={node_name}").items

        total_cpu_limits = 0
        total_memory_limits = 0

        for pod in pods:
            for container in pod.spec.containers:
                resources = container.resources
                if resources.limits:
                    total_cpu_limits += parse_cpu_value(resources.limits.get("cpu", "0"))
                    total_memory_limits += parse_memory_value(resources.limits.get("memory", "0"))
                    

        total_cpu_utilization += total_cpu_limits
        total_memory_utilization += total_memory_limits

     average_cpu_utilization = total_cpu_utilization / total_node_cpu
     average_memory_utilization = total_memory_utilization / total_node_memory
     return {"cpu": round(average_cpu_utilization, 3)*100, "memory": round(average_memory_utilization,3)*100}



def scan_cluster():
    client = OpenshiftClient()
    pods = client.get_resources("Pod")
    instances = {}
    for pod in pods.items:
        for container in pod["spec"]["containers"]:
            resources = get_container_resources(container)
            if not resources:
                continue
            size = get_instance_size(resources)
            if size in instances.keys():
                instances[size] +=1
            else:
                instances[size] =1 
    nodes = client.get_resources("Node")
    nodes_values = calculate_node_avg_allocation(nodes, client)
    nodes_values["number"] = len(nodes.items)
    write_to_csv_file(instances, nodes_values)
     
if __name__ == "__main__":
    scan_cluster()

        






