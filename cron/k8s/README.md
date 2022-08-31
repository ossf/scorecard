# Applying changes to the `openssf` cluster

Currently there is no automation to sync changes to these files to the GKE cluster.
Changes must be manually applied with `kubectl` by a user with permissions to modify the cluster.

## Installing `kubectl`

Follow instructions
[here](https://cloud.google.com/kubernetes-engine/docs/how-to/cluster-access-for-kubectl)
to configure `kubectl` and set the deafult cluster.

The cluster name is `openssf` which is in zone `us-central1-c`.

## Uploading a configuration file

1. Verify you're working on the `openssf` cluster with `kubectl config current-context`
2. Run `kubectl apply -f FILENAME` to apply a new configuration
