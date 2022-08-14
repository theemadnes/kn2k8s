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