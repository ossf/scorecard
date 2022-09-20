# Applying changes to the `openssf` cluster

Currently there is no automation to sync changes to these files to the GKE cluster.
Changes must be manually applied with `kubectl` by a user with permissions to modify the cluster.

## Installing `kubectl`

Follow instructions
[here](https://cloud.google.com/kubernetes-engine/docs/how-to/cluster-access-for-kubectl)
to configure `kubectl` and set the deafult cluster.

The cluster name is `openssf` which is in zone `us-central1-c`.

## Uploading a cronjob/pod configuration file

1. Verify you're working on the `openssf` cluster with `kubectl config current-context`
2. Run `kubectl apply -f FILENAME` to apply a new configuration


## Creating or updating the ConfigMap using the config.yaml file

We use [ConfigMaps](https://kubernetes.io/docs/concepts/configuration/configmap/) to store our config file (`cron/internal/config/config.yaml`).
The file can be created for the first time, or updated, with the same command:
```
kubectl create configmap scorecard-config --from-file=config.yaml -o yaml --dry-run=client | kubectl apply -f -
```

### Accessing the config.yaml through ConfigMap 
The ConfigMap is then volume mounted, so the config file is accessible by any cronjob that specifies the mounting in its yaml.
