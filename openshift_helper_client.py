from kubernetes import config, client
from openshift.dynamic import DynamicClient

ApiVersions = {
        "Deployment": "apps/v1",
        "DeploymentConfig": "apps.openshift.io/v1",
        "StatefulSet": "apps/v1",
        "Pod":"v1",
        "Namespace": "v1",
        "Node": "v1"
    }
    

class OpenshiftClient:
    

    def __init__(self) -> None:
        self.k8s_client = config.new_client_from_config()
        dyn_client = DynamicClient(self.k8s_client)
        self.client = dyn_client


    def get_resources(self, kind: str, namespace: str = None, field_selector: str = None):
        resource_client = self.client.resources.get(api_version=ApiVersions[kind], kind=kind)
        if namespace:
            return resource_client.get(field_selector=field_selector, namespace=namespace)
        return resource_client.get(field_selector=field_selector)
    