package main

import (
	"bufio"
	"bytes"
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

func main() {

	info, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if info.Mode()&os.ModeCharDevice != 0 || info.Size() <= 0 {
		fmt.Println("The command is intended to work with pipes.")
		fmt.Println("Usage: cat source_revision.yaml | kn2k8s")
		return
	}

	reader := bufio.NewReader(os.Stdin)
	var output []uint8

	for {
		input, err := reader.ReadByte()
		if err != nil && err == io.EOF {
			break
		}
		output = append(output, input)
	}

	fmt.Println(generateDeploymentSpec(output))

}
