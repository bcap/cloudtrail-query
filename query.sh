#!/bin/bash

set -e

docker run -i --rm -v ~/.aws:/root/.aws bcap/cloudtrail-query:latest "$@"