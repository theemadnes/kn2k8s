package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"

	k8Yaml "k8s.io/apimachinery/pkg/util/yaml"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	knative "knative.dev/serving/pkg/apis/serving/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	Yml "sigs.k8s.io/yaml"
)

func generateDeploymentSpec(stream []uint8) string {

	rev := &knative.Revision{}
	pod_1 := &corev1.Pod{}
	dep_1 := &appsv1.Deployment{}
	var dep_replicas int32 = 1 // default number of replicas for the generated deployment

	dec := k8Yaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(stream)), 1000)

	if err := dec.Decode(&rev); err != nil {
		fmt.Printf("error decoding the yaml: %v", err)
	}

	// copy podspec from YAML import to new pod
	pod_1.Spec = rev.DeepCopy().Spec.PodSpec

	// get and set service name from revision to pod
	s_name := rev.Labels["serving.knative.dev/service"]
	pod_1.Name = s_name
	pod_1.Spec.Containers[0].Name = s_name

	// hard code the apiVersion & kind
	dep_1.APIVersion = "apps/v1"
	dep_1.Kind = "Deployment"

	// set deployment name and replica count
	dep_1.Name = s_name
	dep_1.Spec.Replicas = &dep_replicas

	// selector & labeling setup for deployment
	app_label_selector_v1 := v1.LabelSelector{}
	label_map := make(map[string]string)
	app_label_selector_v1.MatchLabels = label_map

	dep_1.Spec.Selector = &app_label_selector_v1

	dep_1.Spec.Template.ObjectMeta.Labels = label_map
	dep_1.Spec.Template.ObjectMeta.Labels["app"] = s_name

	// create some annotations for the deployment's template metadata that indicate the source kn revision
	annotation_map := make(map[string]string)
	annotation_map["sourceKnativeService"] = s_name
	annotation_map["sourceKnativeRevision"] = rev.ObjectMeta.Name
	annotation_map["sourceKnativeServiceAccount"] = rev.Spec.ServiceAccountName
	dep_1.Spec.Template.Annotations = annotation_map

	// copy the podspec container spec to the deployment container spec
	dep_1.Spec.Template.Spec.Containers = pod_1.Spec.Containers

	dep_1_yaml, err := Yml.Marshal(dep_1)

	if err != nil {
		fmt.Printf("err: %v\n", err)
		return err.Error()
	}

	return string(dep_1_yaml)
}

func generateServiceSpec(stream []uint8, serviceType string, servicePort int) string {

	service_1 := &corev1.Service{}
	rev := &knative.Revision{}
	service_port := &corev1.ServicePort{}

	dec := k8Yaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(stream)), 1000)

	if err := dec.Decode(&rev); err != nil {
		fmt.Printf("error decoding the yaml: %v", err)
	}

	/*fmt.Println(serviceType)
	fmt.Println(servicePort)*/

	// set Kind & API
	service_1.Kind = "Service"
	service_1.APIVersion = "v1"
	service_1.Spec.Type = corev1.ServiceType(serviceType)
	service_port.Name = rev.Spec.Containers[0].Ports[0].Name
	service_port.Protocol = rev.Spec.Containers[0].Ports[0].Protocol
	service_port.Port = int32(servicePort)
	service_port.TargetPort.IntVal = rev.Spec.Containers[0].Ports[0].ContainerPort
	service_1.Spec.Ports = append(service_1.Spec.Ports, *service_port)
	label_map := make(map[string]string)
	label_map["app"] = rev.Labels["serving.knative.dev/service"]
	service_1.Spec.Selector = label_map

	// set Service name
	service_1.Name = rev.Labels["serving.knative.dev/service"]

	service_1_yaml, err := Yml.Marshal(service_1)

	if err != nil {
		fmt.Printf("err: %v\n", err)
		return err.Error()
	}

	return string(service_1_yaml)
}

func main() {

	// pull optional command line params (used to configure service port & service type)
	serviceTypePtr := flag.String("serviceType", "ClusterIP", "string to indicate type of service to create")
	servicePortPtr := flag.Int("servicePort", 80, "unsigned int to set external port used by service")

	flag.Parse()

	// set up piping stdin to utility
	reader := bufio.NewReader(os.Stdin)
	var output []uint8

	for {
		input, err := reader.ReadByte()
		if err != nil && err == io.EOF {
			break
		}
		output = append(output, input)
	}

	// generate deployment YAML
	fmt.Println(generateDeploymentSpec(output))

	// add multi-resource delimeter
	fmt.Println("---")

	// generate service YAML
	fmt.Println(generateServiceSpec(output, *serviceTypePtr, *servicePortPtr))

}
