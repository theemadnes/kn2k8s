# kn2k8s

command line tool for converting knative revision spec YAML to Kubernetes API & Gateway API specs. 

The script expects input via stdin, so you'll need to pass the input via a pipe.

It'll create a namespace, deployment, service account, service account, HPA, HTTPRoute, etc based on the source service name.

#### usage

Apply a locally stored knative revision spec:

```cat samples/hello_00005_qud.yaml | go run . | kubectl apply -f -```

*or* apply a spec from Cloud Run (managed), which takes a little while to get the YAML via `gcloud` CLI:

```gcloud run revisions describe hello-00002-muc --region=us-central1  --project am-arg-01 --format yaml | go run . | kubectl apply -f -```

Optionally can override service defaults (service type of `ClusterIP` on port `80`) via CLI flags:

```cat samples/hello_00005_qud.yaml | go run . -serviceType=LoadBalancer -servicePort=8080 | kubectl apply -f -```

And HPA max/min replica defaults (minReplicas of `1` and maxReplicas from revision spec `maxReplicas`):

```cat samples/hello_00005_qud.yaml | go run . -maxReplicas=10 -minReplicas=2 -serviceType=LoadBalancer | kubectl apply -f -```

Or create a bunch of minReplicas:

```cat samples/hello_00005_qud.yaml | go run . -maxReplicas=50 -minReplicas=10 | kubectl apply -f - ```