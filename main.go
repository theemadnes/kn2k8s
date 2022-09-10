package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
	k8Yaml "k8s.io/apimachinery/pkg/util/yaml"

	appsv1 "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	knative "knative.dev/serving/pkg/apis/serving/v1"

	//"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/schollz/progressbar/v3"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gw "sigs.k8s.io/gateway-api/apis/v1beta1"
	Yml "sigs.k8s.io/yaml"
)

// struct for unmarshalling revision details from command line
type Revisions struct {
	Revision []struct {
		RevisionId string `yaml:"revision_id"`
		Region     string `yaml:"region"`
		ProjectId  string `yaml:"project_id"`
	} `yaml:"revisions"`
}

// function to read revision manifest from command line
func readRevisions(filename string) (Revisions, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	//r := &revisions{}
	r := Revisions{}
	err = yaml.Unmarshal(buf, &r)
	if err != nil {
		return r, fmt.Errorf("in file %q: %v", filename, err)
	}

	return r, nil
}

// super hacky way of removing empty fields
// the benefit of doing this is that resources don't always show as `configured` even when there are no changes
func hackToRemoveEmptyFields(ymlBytes []byte) []byte {
	lines := bytes.Replace(ymlBytes, []byte("      creationTimestamp: null\n"), []byte(""), 1)
	lines = bytes.Replace(lines, []byte("spec: {}\n"), []byte(""), 1)
	lines = bytes.Replace(lines, []byte("  creationTimestamp: null\n"), []byte(""), -1)
	lines = bytes.Replace(lines, []byte("status: {}\n"), []byte(""), 1)
	lines = bytes.Replace(lines, []byte("status:\n"), []byte(""), 1)
	lines = bytes.Replace(lines, []byte("  strategy: {}\n"), []byte(""), 1)
	lines = bytes.Replace(lines, []byte("  loadBalancer: {}\n"), []byte(""), 1)
	lines = bytes.Replace(lines, []byte("currentMetrics: null\n"), []byte(""), 1)
	lines = bytes.Replace(lines, []byte("    desiredReplicas: 0\n"), []byte(""), 1)
	lines = bytes.Replace(lines, []byte("  parents: null\n"), []byte(""), 1)
	return lines
}

func generateNamespaceSpec(stream []uint8) []byte {

	rev := &knative.Revision{}
	ns_1 := &corev1.Namespace{}

	dec := k8Yaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(stream)), 1000)

	if err := dec.Decode(&rev); err != nil {
		log.Fatal(err)
	}

	ns_1.APIVersion = "v1"
	ns_1.Kind = "Namespace"
	ns_1.ObjectMeta.Name = rev.Labels["serving.knative.dev/service"]

	ns_1_yaml, err := Yml.Marshal(ns_1)

	if err != nil {
		log.Fatal(err)
	}

	ns_1_yaml = hackToRemoveEmptyFields(ns_1_yaml)

	return ns_1_yaml

}

func generateServiceAccountSpec(stream []uint8) []byte {

	rev := &knative.Revision{}
	sa_1 := &corev1.ServiceAccount{}

	dec := k8Yaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(stream)), 1000)

	if err := dec.Decode(&rev); err != nil {
		log.Fatal(err)
	}

	// set up basics for service account
	sa_1.APIVersion = "v1"
	sa_1.Kind = "ServiceAccount"
	sa_1.ObjectMeta.Name = rev.Labels["serving.knative.dev/service"]
	sa_1.ObjectMeta.Namespace = rev.Labels["serving.knative.dev/service"]

	// configure the KSA to be ready for Workload Identity
	sa_1.Annotations = make(map[string]string)
	sa_1.Annotations["iam.gke.io/gcp-service-account"] = rev.Spec.ServiceAccountName

	sa_1_yaml, err := Yml.Marshal(sa_1)

	if err != nil {
		log.Fatal(err)
	}

	sa_1_yaml = hackToRemoveEmptyFields(sa_1_yaml)

	return sa_1_yaml
}

func generateDeploymentSpec(stream []uint8) []byte {

	rev := &knative.Revision{}
	pod_1 := &corev1.Pod{}
	dep_1 := &appsv1.Deployment{}
	var dep_replicas int32 = 1 // default number of replicas for the generated deployment

	dec := k8Yaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(stream)), 1000)

	if err := dec.Decode(&rev); err != nil {
		log.Fatal(err)
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
	dep_1.ObjectMeta.Namespace = rev.Labels["serving.knative.dev/service"]

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

	// set the service account
	dep_1.Spec.Template.Spec.ServiceAccountName = s_name

	dep_1_yaml, err := Yml.Marshal(dep_1)

	if err != nil {
		log.Fatal(err)
	}

	dep_1_yaml = hackToRemoveEmptyFields(dep_1_yaml)

	return dep_1_yaml
}

func generateServiceSpec(stream []uint8, serviceType string, servicePort int) []byte {

	service_1 := &corev1.Service{}
	rev := &knative.Revision{}
	service_port := &corev1.ServicePort{}

	dec := k8Yaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(stream)), 1000)

	if err := dec.Decode(&rev); err != nil {
		log.Fatal(err)
	}

	// set Kind & API
	service_1.Kind = "Service"
	service_1.APIVersion = "v1"
	service_1.Spec.Type = corev1.ServiceType(serviceType)
	service_1.ObjectMeta.Namespace = rev.Labels["serving.knative.dev/service"]
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
		log.Fatal(err)
	}

	service_1_yaml = hackToRemoveEmptyFields(service_1_yaml)

	return service_1_yaml
}

func generateHorizontalPodAutoscalerSpec(stream []uint8, minReplicas int, maxReplicas int) []byte {

	// create HPA resource
	hpa_1 := &autoscaling.HorizontalPodAutoscaler{}
	rev := &knative.Revision{}

	dec := k8Yaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(stream)), 1000)

	if err := dec.Decode(&rev); err != nil {
		log.Fatal(err)
	}

	// set prelim vars
	var avgUtilization int32 = 50
	var currentReplicas int32 = 1
	var minReplicasPtr *int32 = new(int32)
	*minReplicasPtr = int32(minReplicas)

	// figure out if maxReplicas has been provided via CLI
	if maxReplicas == 0 {
		maxReplicas, _ = strconv.Atoi(rev.Annotations["autoscaling.knative.dev/maxScale"])
	}

	// define fields
	hpa_1.APIVersion = "autoscaling/v2beta2"
	hpa_1.Kind = "HorizontalPodAutoscaler"
	hpa_1.ObjectMeta.Name = rev.Labels["serving.knative.dev/service"]
	hpa_1.ObjectMeta.Namespace = rev.Labels["serving.knative.dev/service"]
	hpa_1.Spec.ScaleTargetRef.APIVersion = "apps/v1"
	hpa_1.Spec.ScaleTargetRef.Kind = "Deployment"
	hpa_1.Spec.ScaleTargetRef.Name = rev.Labels["serving.knative.dev/service"]
	hpa_1.Spec.MinReplicas = minReplicasPtr
	hpa_1.Spec.MaxReplicas = int32(maxReplicas)
	hpa_1.Status = autoscaling.HorizontalPodAutoscalerStatus{}
	hpa_1.Status.CurrentReplicas = currentReplicas
	hpa_1.Status.Conditions = append(hpa_1.Status.Conditions, autoscaling.HorizontalPodAutoscalerCondition{})

	// create CPU-based metric spec
	metricsSpec := &autoscaling.MetricSpec{}
	metricsResource := &autoscaling.ResourceMetricSource{}
	metricsResource.Name = "cpu"
	metricsResource.Target.Type = "Utilization"
	metricsResource.Target.AverageUtilization = &avgUtilization
	metricsSpec.Type = "Resource"
	metricsSpec.Resource = metricsResource

	hpa_1.Spec.Metrics = append(hpa_1.Spec.Metrics, *metricsSpec)
	hpa_1.Status = autoscaling.HorizontalPodAutoscaler{}.Status // set blank status for easier cleanup

	hpa_1_yaml, err := Yml.Marshal(hpa_1)

	if err != nil {
		log.Fatal(err)
	}

	hpa_1_yaml = hackToRemoveEmptyFields(hpa_1_yaml)

	return hpa_1_yaml
}

func generateHttpRouteSpec(stream []uint8, gwName string, gwNamespace string, svcPort int) []byte {

	httpRoute_1 := gw.HTTPRoute{}
	httpRouteParentRef := gw.ParentReference{}
	httpRouteName := gw.ObjectName(gwName)
	httpRouteNamespace := gw.Namespace(gwNamespace)
	httpRouteRulesPathMatch := gw.HTTPPathMatch{}
	httpRoutePathMatchType := gw.PathMatchType("Exact")
	httpRouteBackendRef := gw.HTTPBackendRef{}
	httpRouteBackendPort := gw.PortNumber(svcPort)

	rev := &knative.Revision{}

	dec := k8Yaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(stream)), 1000)

	if err := dec.Decode(&rev); err != nil {
		log.Fatal(err)
	}

	// set path details
	// httpRouteRulesPath := gw.PathMatchPathPrefix("/" + rev.Labels["serving.knative.dev/service"])
	prefix := "/" + rev.Labels["serving.knative.dev/service"]
	httpRouteRulesPathMatch.Value = &prefix
	httpRouteRulesPathMatch.Type = &httpRoutePathMatchType
	httpRouteBackendRef.Name = gw.ObjectName(rev.Labels["serving.knative.dev/service"])
	httpRouteBackendRef.Port = &httpRouteBackendPort

	// set API version and Kind
	httpRoute_1.APIVersion = "gateway.networking.k8s.io/v1beta1"
	httpRoute_1.Kind = "HTTPRoute"

	// configure metadata
	httpRoute_1.ObjectMeta.Name = rev.Labels["serving.knative.dev/service"]
	httpRoute_1.ObjectMeta.Namespace = rev.Labels["serving.knative.dev/service"]

	// configure nested fields
	httpRouteParentRef.Name = httpRouteName
	httpRouteParentRef.Namespace = &httpRouteNamespace
	httpRoute_1.Spec.ParentRefs = make([]gw.ParentReference, 1)
	httpRoute_1.Spec.ParentRefs[0] = httpRouteParentRef
	httpRoute_1.Spec.Rules = make([]gw.HTTPRouteRule, 1)
	//httpRoute_1.Spec.Rules[0] = httpRouteRules
	httpRoute_1.Spec.Rules[0].Matches = make([]gw.HTTPRouteMatch, 1)
	httpRoute_1.Spec.Rules[0].Matches[0].Path = &httpRouteRulesPathMatch
	httpRoute_1.Spec.Rules[0].BackendRefs = make([]gw.HTTPBackendRef, 1)
	httpRoute_1.Spec.Rules[0].BackendRefs[0] = httpRouteBackendRef

	httpRoute_1_yaml, err := Yml.Marshal(httpRoute_1)

	if err != nil {
		log.Fatal(err)
	}

	// clean up HTTPRoute output
	httpRoute_1_yaml = hackToRemoveEmptyFields(httpRoute_1_yaml)

	return httpRoute_1_yaml
}

func kubectlApply(path string) {
	var kubectl_out bytes.Buffer
	kubectl_cmd := exec.Command("kubectl", "apply", "-f", path)
	//fmt.Println(kubectl_cmd)
	kubectl_cmd.Stdout = &kubectl_out
	kubectl_err := kubectl_cmd.Run()
	if kubectl_err != nil {
		log.Fatal(kubectl_err)
	}
	//fmt.Printf(kubectl_out.String())
}

func getServiceInfo(stream []uint8) map[string]string {

	revInfo := make(map[string]string)

	rev := &knative.Revision{}

	dec := k8Yaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(stream)), 1000)

	if err := dec.Decode(&rev); err != nil {
		log.Fatal(err)
	}

	revInfo["serviceName"] = rev.Labels["serving.knative.dev/service"]
	revInfo["revisionId"] = rev.Name
	revInfo["image"] = rev.Spec.Containers[0].Image
	revInfo["cpu"] = rev.Spec.Containers[0].Resources.Limits.Cpu().String()
	revInfo["memory"] = rev.Spec.Containers[0].Resources.Limits.Memory().String()

	return revInfo
}

func main() {

	// pull optional command line params (used to configure service port & service type)

	manifestFilePtr := flag.String("manifestFile", "manifest.yaml", "input file name")
	serviceTypePtr := flag.String("serviceType", "ClusterIP", "string to indicate type of service to create")
	servicePortPtr := flag.Int("servicePort", 80, "int to set external port used by service")
	maxReplicasPtr := flag.Int("maxReplicas", 0, "int to set maximum replicas via HPA - otherwise will set to revision maxScale value") // default to zero to detect input
	minReplicasPtr := flag.Int("minReplicas", 1, "int to set minimum replicas via HPA")
	gwNamePtr := flag.String("gatewayName", "external-http", "string of gateway object name")
	gwNamespacePtr := flag.String("gatewayNamespace", "external-gw", "string of gateway namespace")

	flag.Parse()

	/*
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
	*/

	// get timestamp
	timeString := string(time.Now().Format(time.RFC3339))
	timeString = strings.Replace(timeString, ":", "", -1)

	fmt.Printf("Reading manifest file %v\n", *manifestFilePtr)

	r, err := readRevisions(*manifestFilePtr)
	if err != nil {
		log.Fatal(err)
	}

	// create output table
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Service", "Revision", "Image", "CPU", "Memory", "Path"})

	// create progress
	//p := progress.NewWriter()
	//p.SetOutputWriter(os.Stdout)

	// cycle through revisions and process
	for _, revision := range r.Revision {

		// get revision info via gcloud
		var out bytes.Buffer
		cmd := exec.Command("gcloud", "run", "revisions", "describe", revision.RevisionId, "--region="+revision.Region, "--project="+revision.ProjectId, "--format=yaml")
		cmd.Stdout = &out
		cmd_err := cmd.Run()

		if err != nil {
			log.Fatal(cmd_err)
		}

		// get revision data
		serviceInfo := getServiceInfo(out.Bytes())

		// create table output
		t.AppendRow([]interface{}{serviceInfo["serviceName"], serviceInfo["revisionId"], serviceInfo["image"], serviceInfo["cpu"], serviceInfo["memory"], "/" + serviceInfo["serviceName"]})
		t.AppendSeparator()

		// set up progress bar
		fmt.Printf("\nProcessing revision %v of service %v\n", serviceInfo["revisionId"], serviceInfo["serviceName"])
		bar := progressbar.NewOptions(100, progressbar.OptionSetWidth(50), progressbar.OptionEnableColorCodes(true), progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
		//bar = bar

		// create subfolder per revision
		pathPrefix := "output/" + timeString + "/" + serviceInfo["serviceName"] + "/"
		err := os.MkdirAll(pathPrefix+"", 0755)
		if err != nil {
			log.Fatal(err)
		}
		bar.Add(10)

		// create and apply YAML files
		ns_err := os.WriteFile(pathPrefix+"ns.yaml", generateNamespaceSpec(out.Bytes()), 0755)
		if ns_err != nil {
			log.Fatal(ns_err)
		}
		kubectlApply(pathPrefix + "ns.yaml")
		bar.Add(15)

		sa_err := os.WriteFile(pathPrefix+"sa.yaml", generateServiceAccountSpec(out.Bytes()), 0755)
		if sa_err != nil {
			log.Fatal(sa_err)
		}
		kubectlApply(pathPrefix + "sa.yaml")
		bar.Add(15)

		deployment_err := os.WriteFile(pathPrefix+"deployment.yaml", generateDeploymentSpec(out.Bytes()), 0755)
		if deployment_err != nil {
			log.Fatal(deployment_err)
		}
		kubectlApply(pathPrefix + "deployment.yaml")
		bar.Add(15)

		service_err := os.WriteFile(pathPrefix+"service.yaml", generateServiceSpec(out.Bytes(), *serviceTypePtr, *servicePortPtr), 0755)
		if service_err != nil {
			log.Fatal(service_err)
		}
		kubectlApply(pathPrefix + "service.yaml")
		bar.Add(15)

		hpa_err := os.WriteFile(pathPrefix+"hpa.yaml", generateHorizontalPodAutoscalerSpec(out.Bytes(), *minReplicasPtr, *maxReplicasPtr), 0755)
		if hpa_err != nil {
			log.Fatal(hpa_err)
		}
		kubectlApply(pathPrefix + "hpa.yaml")
		bar.Add(15)

		route_err := os.WriteFile(pathPrefix+"route.yaml", generateHttpRouteSpec(out.Bytes(), *gwNamePtr, *gwNamespacePtr, 80), 0755)
		if route_err != nil {
			log.Fatal(route_err)
		}
		kubectlApply(pathPrefix + "route.yaml")
		bar.Add(15)
	}

	fmt.Printf("\nResults:\n")
	t.Render()
}
