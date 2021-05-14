# Generating proto files

## Installation

Follow instructions
[here](https://developers.google.com/protocol-buffers/docs/gotutorial#compiling-your-protocol-buffers)
to install necessary binaries.

## Compile

Use the command below to compile:

```
protoc --go_out=$DST_DIR request.proto
```

NOTE: $DST_DIR should contain `github.com/ossf/scorecard/cron/data` directory
structure.

## Future work

Update Makefile to compile and generate proto files, when we run `make all`.
