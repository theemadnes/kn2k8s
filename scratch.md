- get examples of knative services 
- check out auth 
- review knative serving API
- install knative on GKE cluster
- show examples of creating and viewing services via gcloud CLI

sudo tcpdump -i any host 10.254.211.134

$ gcloud --project am-arg-01 run services list
   SERVICE  REGION       URL                                    LAST DEPLOYED BY                 LAST DEPLOYED AT
✔  hello    us-central1  https://hello-4uotx33u2a-uc.a.run.app  admin@alexmattson.altostrat.com  2022-07-27T05:11:34.852305Z


$ gcloud --project am-arg-01 run services describe hello
✔ Service hello in region us-central1
 
URL:     https://hello-4uotx33u2a-uc.a.run.app
Ingress: all
Traffic:
  100% LATEST (currently hello-00001-rec)
 
Last updated on 2022-07-27T05:11:34.852305Z by admin@alexmattson.altostrat.com:
  Revision hello-00001-rec
  Image:           us-docker.pkg.dev/cloudrun/container/hello
  Port:            8080
  Memory:          512Mi
  CPU:             1000m
  Service account: 841101411908-compute@developer.gserviceaccount.com
  Concurrency:     80
  Max Instances:   100
  Timeout:         300s

$ gcloud --project am-arg-01 run revisions describe hello-00001-rec
✔ Revision hello-00001-rec in region us-central1
 
Image:           us-docker.pkg.dev/cloudrun/container/hello@sha256:717e538e1ef8f955a54834e213d080bde6a8b3513fcc406df0d5d5ed3ed2853b
Port:            8080
Memory:          512Mi
CPU:             1000m
Service account: 841101411908-compute@developer.gserviceaccount.com
Concurrency:     80
Max Instances:   100
Timeout:         300s
CPU Allocation:  CPU is only allocated during request processing


https://knative.dev/docs/serving/#serving-resources

- service
- revision
  - PIT of code
- configuration
- route
  - fractional traffic 


https://ahmet.im/blog/cloud-run-is-a-knative/

Service or ingress
TLS 
Service account 

$ gcloud --project am-arg-01 run revisions list
   REVISION         ACTIVE  SERVICE  DEPLOYED                 DEPLOYED BY
✔  hello-00002-muc          hello    2022-07-29 04:38:51 UTC  admin@alexmattson.altostrat.com
✔  hello-00001-rec  yes     hello    2022-07-27 05:10:11 UTC  admin@alexmattson.altostrat.com

$ gcloud --project am-arg-01 run services describe hello
✔ Service hello in region us-central1
 
URL:     https://hello-4uotx33u2a-uc.a.run.app
Ingress: all
Traffic:
  50% hello-00001-rec
  50% hello-00002-muc
 
Last updated on 2022-07-29T04:40:18.758278Z by admin@alexmattson.altostrat.com:
  Revision hello-00002-muc
  Image:           us-docker.pkg.dev/google-samples/containers/gke/whereami:v1.2.9
  Port:            8080
  Memory:          512Mi
  CPU:             1000m
  Service account: 841101411908-compute@developer.gserviceaccount.com
  Concurrency:     80
  Max Instances:   100
  Timeout:         300s


go mod init theemades/kn2k8s
go mod tidy


$ gcloud run revisions list --service=hello --project am-arg-01
   REVISION         ACTIVE  SERVICE  DEPLOYED                 DEPLOYED BY
✔  hello-00002-muc  yes     hello    2022-07-29 04:38:51 UTC  admin@alexmattson.altostrat.com
✔  hello-00001-rec  yes     hello    2022-07-27 05:10:11 UTC  admin@alexmattson.altostrat.com

$ gcloud run revisions describe hello-00002-muc --region=us-central1  --project am-arg-01
✔ Revision hello-00002-muc in region us-central1
 
Image:           us-docker.pkg.dev/google-samples/containers/gke/whereami@sha256:9957f5ff3096a83bae4e0952faaebcac740557e7fb2a642ed38bf5cb64c45795
Port:            8080
Memory:          512Mi
CPU:             1000m
Service account: 841101411908-compute@developer.gserviceaccount.com
Concurrency:     80
Max Instances:   100
Timeout:         300s
CPU Allocation:  CPU is only allocated during request processing


go get gopkg.in/yaml.v3

https://github.com/knative/specs/blob/main/specs/serving/overview.md#configuration

https://pkg.go.dev/knative.dev/serving/pkg/apis/serving/v1#RevisionSpec

ishmeet
go get k8s.io/apimachinery/pkg/util/yaml
go get knative.dev/serving/pkg/apis/serving/v1

go get k8s.io/api/core/v1
go get k8s.io/api/apps/v1
go get sigs.k8s.io/yaml
go get k8s.io/apimachinery
go get k8s.io/apimachinery/pkg/labels


$ gcloud run revisions describe hello-00002-muc --region=us-central1  --project am-arg-01 --format yaml
apiVersion: serving.knative.dev/v1
kind: Revision
metadata:
  annotations:
    autoscaling.knative.dev/maxScale: '100'
    run.googleapis.com/client-name: cloud-console
    run.googleapis.com/cpu-throttling: 'true'
    serving.knative.dev/creator: admin@alexmattson.altostrat.com
  creationTimestamp: '2022-07-29T04:38:51.599559Z'
  generation: 1
  labels:
    cloud.googleapis.com/location: us-central1
    serving.knative.dev/configuration: hello
    serving.knative.dev/configurationGeneration: '2'
    serving.knative.dev/route: hello
    serving.knative.dev/service: hello
    serving.knative.dev/serviceUid: 4c2ce787-6f11-4a7c-b101-04e486d26c4a
  name: hello-00002-muc
  namespace: '841101411908'
  ownerReferences:
  - apiVersion: serving.knative.dev/v1
    blockOwnerDeletion: true
    controller: true
    kind: Configuration
    name: hello
    uid: 74f95133-bc6a-462f-87af-802d6baee0f3
  resourceVersion: AAXk6mLTQk4
  selfLink: /apis/serving.knative.dev/v1/namespaces/841101411908/revisions/hello-00002-muc
  uid: 8072d5f3-cdf4-4f22-840b-690f7559cd64
spec:
  containerConcurrency: 80
  containers:
  - image: us-docker.pkg.dev/google-samples/containers/gke/whereami@sha256:9957f5ff3096a83bae4e0952faaebcac740557e7fb2a642ed38bf5cb64c45795
    ports:
    - containerPort: 8080
      name: http1
    resources:
      limits:
        cpu: 1000m
        memory: 512Mi
  serviceAccountName: 841101411908-compute@developer.gserviceaccount.com
  timeoutSeconds: 300
status:
  conditions:
  - lastTransitionTime: '2022-07-29T04:39:03.722467Z'
    status: 'True'
    type: Ready
  - lastTransitionTime: '2022-07-29T04:40:18.676850Z'
    severity: Info
    status: 'True'
    type: Active
  - lastTransitionTime: '2022-07-29T04:39:03.722467Z'
    status: 'True'
    type: ContainerHealthy
  - lastTransitionTime: '2022-07-29T04:39:00.654502Z'
    status: 'True'
    type: ResourcesAvailable
  imageDigest: us-docker.pkg.dev/google-samples/containers/gke/whereami@sha256:9957f5ff3096a83bae4e0952faaebcac740557e7fb2a642ed38bf5cb64c45795
  logUrl: https://console.cloud.google.com/logs/viewer?project=am-arg-01&resource=cloud_run_revision/service_name/hello/revision_name/hello-00002-muc&advancedFilter=resource.type%3D%22cloud_run_revision%22%0Aresource.labels.service_name%3D%22hello%22%0Aresource.labels.revision_name%3D%22hello-00002-muc%22
  observedGeneration: 1