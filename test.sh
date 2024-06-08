#!/bin/sh

docker image build --tag go-test-messages --file Dockerfile.test . 
docker container run --rm go-test-messages:latest
exit_code=$?
docker image rm -f go-test-messages

exit $exit_code
