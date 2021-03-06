# Deploying within k8s

The gitcache is deployed within k8s using `deployment.yaml`. The gitcache service isn't exposed as external service.

The gitcache would authenticate with `bucket` using https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity#gcloud in GCP.

