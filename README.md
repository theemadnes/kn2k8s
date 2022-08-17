# kn2k8s
command line tool for converting knative revision spec YAML to Kubernetes API specs

#### usage

```gcloud run revisions describe hello-00002-muc --region=us-central1  --project am-arg-01 --format yaml | go run .```

*or*

```cat samples/hello_revision_4.yaml | go run . | kubectl apply -f -```