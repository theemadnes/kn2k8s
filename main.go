package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"reflect"

	k8Yaml "k8s.io/apimachinery/pkg/util/yaml"

	//v1 "knative.dev/pkg/apis/duck/v1"
	knative "knative.dev/serving/pkg/apis/serving/v1"
	//k8Api "k8s.io/kubernetes/pkg/api/v1"
	//k8Api "k8s.io/api"
	//api "k8s.io/kubernetes/pkg/api"
	//resource "k8s.io/kubernetes/pkg/api/resource"
	//metav1 "k8s.io/api/meta/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	//lz "k8s.io/apimachinery/pkg/labels"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
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

	//fmt.Printf("%+v", *rev)
	//fmt.Println(reflect.TypeOf(*rev))
	//fmt.Println(rev.TypeMeta.Kind)
	//fmt.Println(rev.ObjectMeta.Name)
	//fmt.Println(rev.Spec.PodSpec)
	//fmt.Println(reflect.TypeOf(rev.Spec.PodSpec))
	// set pod name from revision's parent service name
	//pod_1.Name = rev.Labels['serving.knative.dev/service']
	//pod_1.Name := rev.Labels["serving.knative.dev/service"].(string)
	/*
		for key, value := range rev.Labels {
			fmt.Println("Key:", key, "Value:", value)
		}*/
	s_name := rev.Labels["serving.knative.dev/service"]
	//fmt.Println(s_name)

	pod_1.Name = s_name
	pod_1.Spec.Containers[0].Name = s_name
	//fmt.Println(pod_1.Name)
	//fmt.Println(pod_1)

	// convert new pod to YAML
	/*
		pod_1_yaml, err := Yml.Marshal(pod_1)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			return
		}
		//fmt.Println(string(pod_1_yaml))*/

	// hard code the apiVersion & kind
	dep_1.APIVersion = "apps/v1"
	dep_1.Kind = "Deployment"

	// set name and replica count
	dep_1.Name = s_name
	dep_1.Spec.Replicas = &dep_replicas

	// selector & labeling setup for deployment
	app_label_requirement, _ := labels.NewRequirement("app", selection.Equals, []string{s_name})
	//app_label_requirement_v1 := v1.LabelSelectorRequirement("app", selection.Equals, s_name) //{}
	//app_label_requirement_v1 := v1.LabelSelectorRequirement{}
	//app_label_requirement_v1.Key = "app"
	//app_label_requirement_v1.Operator = v1.LabelSelectorOperator(selection.Equals)
	//app_label_requirement_v1.Values = []string{s_name}
	//app_label_requirement_v1.
	app_label_selector := labels.NewSelector()
	app_label_selector_v1 := v1.LabelSelector{}
	label_map := make(map[string]string)
	app_label_selector_v1.MatchLabels = label_map
	//app_label_selector_v1.MatchExpressions = append(app_label_selector_v1.MatchExpressions, app_label_requirement_v1)

	app_label_selector = app_label_selector.Add(*app_label_requirement)

	// set selector
	//label_map := make(map[string]string)
	//println(reflect.TypeOf(&dep_1.Spec.Selector))
	//label_selector, err := lz.NewRequirement("app", selection.Equals, []string{s_name})

	//dep_1.Spec.Selector = v1.SetAsLabelSelector(label_map)
	//dep_1.Spec.Selector.MatchExpressions[0] = v1.LabelSelectorRequirement{}
	dep_1.Spec.Selector = &app_label_selector_v1

	//dep_1.Spec.Selector.MatchLabels = label_map
	//dep_1.Spec.Selector = lbls.NewRequirement("app", selection.Equals, []string{s_name})
	//dep_1.Spec.Selector.MatchExpressions[0] = *label_selector
	//dep_1.Spec.Selector.MatchLabels["app"] = s_name
	//fmt.Println(dep_1.Spec.Template.ObjectMeta.Labels)

	//container_map := make(map[int32]corev1.Container)
	dep_1.Spec.Template.ObjectMeta.Labels = label_map
	dep_1.Spec.Template.ObjectMeta.Labels["app"] = s_name
	//fmt.Println(dep_1.Spec.Template.Spec.Containers)

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

	// spec := &knative.spec{}
	stream, err := ioutil.ReadFile("samples/hello_revision_4.yaml")
	fmt.Println(reflect.TypeOf(stream))
	if err != nil {
		fmt.Println("Cannot open the file", err)
	}

	fmt.Println(generateDeploymentSpec(stream))

	//fmt.Println(dep_1)
}
