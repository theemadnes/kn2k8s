# kn2k8s

command line tool for converting Cloud Run revisions to Kubernetes primitives.

it will read a YAML file as input - example:

```
revisions:
- revision_id: hola-00001-tof
  region: us-central1
  project_id: am-arg-01
- revision_id: hello-00005-qud
  region: us-central1
  project_id: am-arg-01
```

It'll create a namespace, deployment, service account, service account, HPA, HTTPRoute, etc based on the source revision info.

#### usage

```go run . manifest.yaml```