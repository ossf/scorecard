# README

The `cron` job is deployed in `k8s` within `gcp`. The `docker` builds for the `cron` is built by cloudbuild within the `OpenSSF` `gcp` project and is pushed into the `gcr`.


![](https://i.imgur.com/thKKlLJ.png)


The cloudbuild runs based on push to the `main` branch.

The `k8s` cron-job uses the `gcr.io` registry to pull images.
