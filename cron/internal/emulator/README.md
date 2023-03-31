## GCS emulator

Use [fake-gcs-server](https://github.com/fsouza/fake-gcs-server). Install from source or `Releases`:

```
go install github.com/fsouza/fake-gcs-server@latest
```

From the base of the Scorecard git repo:
```
fake-gcs-server -scheme http -public-host 0.0.0.0:4443 \
    -backend filesystem -filesystem-root cron/internal/emulator/mockgcs
```

## pubsub emulator:
Install following directions at https://cloud.google.com/pubsub/docs/emulator


#### Initial python-pubsub setup
```
git clone https://github.com/googleapis/python-pubsub
cd python-pubsub/samples/snippet
pip install -r requirements.txt
```

### Configure Topic and Subscription

#### Env stuff
```
export PUBSUB_PROJECT_ID=test
export TOPIC_ID=scorecard-batch-requests
export SUBSCRIPTION_ID=scorecard-batch-worker
```

1. start emulator
```
export PUBSUB_PROJECT_ID=test
gcloud beta emulators pubsub start --project=$PUBSUB_PROJECT_ID
```

2. setup topics/subs
```
$(gcloud beta emulators pubsub env-init)
python publisher.py $PUBSUB_PROJECT_ID create $TOPIC_ID
python subscriber.py $PUBSUB_PROJECT_ID create $TOPIC_ID $SUBSCRIPTION_ID
```


#### Drain the queue
```
python subscriber.py $PUBSUB_PROJECT_ID receive $SUBSCRIPTION_ID
```

### run scorecard

#### worker
```
$(gcloud beta emulators pubsub env-init)
export STORAGE_EMULATOR_HOST=0.0.0.0:4443
go run cron/internal/worker/!(*_test).go \
    --ignoreRuntimeErrors=true \
    --config cron/internal/emulator/config.yaml
```

#### controller
```
$(gcloud beta emulators pubsub env-init)
export STORAGE_EMULATOR_HOST=0.0.0.0:4443
go run cron/internal/controller/!(*_test).go \
    --config cron/internal/emulator/config.yaml \
    cron/internal/emulator/projects.csv
```
