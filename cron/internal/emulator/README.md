# Configuring a local environment to test the Scorecard Cron Job

This emulator focuses on being able to test the `worker`, which pulls messages from a pubsub, processes them, and writes the results to a Google Cloud Storage (GCS) bucket.
It's necessary to support pubsub, gcs, and the `controller` to get the `worker` working.

In general, you'll need 4-5 terminals (or tmux) to run everything needed.

## GCS emulator

[fake-gcs-server](https://github.com/fsouza/fake-gcs-server) meets our needs and is written in Go. 
We may be able to use it as a library for unit tests in the future.

For now, the binary is good enough, so install it from source (or [Releases](https://github.com/fsouza/fake-gcs-server/releases)):

```
go install github.com/fsouza/fake-gcs-server@latest
```

Now you can run the fake from the root of the Scorecard repo in your first window:
```
fake-gcs-server -scheme http -public-host 0.0.0.0:4443 \
    -backend filesystem -filesystem-root cron/internal/emulator/fakegcs
```

## pubsub emulator:
Google Cloud has a [pubsub emulator](https://cloud.google.com/pubsub/docs/emulator) with complete install ininstructions.
I've summarized some of them below.


### One time setup

```
gcloud components install pubsub-emulator
gcloud components update
```

Anywhere outside your scorecard repo:
```
git clone https://github.com/googleapis/python-pubsub
cd python-pubsub/samples/snippet
pip install -r requirements.txt
```

### Running the pubsub emulator (needed to do everytime)

In a second window from any directory, run the emulator itself:

```
export PUBSUB_PROJECT_ID=test
gcloud beta emulators pubsub start --project=$PUBSUB_PROJECT_ID
```

In a third window (from the `samples/snippet` directory wherever you cloned `python-pubsub`) create the topic and subscription:

```
export PUBSUB_PROJECT_ID=test
export TOPIC_ID=scorecard-batch-requests
export SUBSCRIPTION_ID=scorecard-batch-worker
$(gcloud beta emulators pubsub env-init)
python3 publisher.py $PUBSUB_PROJECT_ID create $TOPIC_ID
python3 subscriber.py $PUBSUB_PROJECT_ID create $TOPIC_ID $SUBSCRIPTION_ID
alias drain-pubsub="python3 subscriber.py $PUBSUB_PROJECT_ID receive $SUBSCRIPTION_ID"
```

At any point you can drain the queue by running the following in the same window. Make sure to stop the command when testing the `worker`:
```
drain-pubsub
```

## run Scorecard cron components

Commands intended to be run from the base of the Scorecard repo. Since this is intended to be used during development, `go run` is used but there's no reason you can't use `go build`. 
The repos in `cron/internal/emulator/projects.csv` and the `cron/internal/emulator/config.yaml` file can be changed as needed.

### controller
```
$(gcloud beta emulators pubsub env-init)
export STORAGE_EMULATOR_HOST=0.0.0.0:4443
go run $(ls cron/internal/controller/*.go | grep -v _test.go) \
    --config cron/internal/emulator/config.yaml \
    cron/internal/emulator/projects.csv
```

### worker
```
$(gcloud beta emulators pubsub env-init)
export STORAGE_EMULATOR_HOST=0.0.0.0:4443
go run $(ls cron/internal/worker/*.go | grep -v _test.go) \
    --ignoreRuntimeErrors=true \
    --config cron/internal/emulator/config.yaml
```
