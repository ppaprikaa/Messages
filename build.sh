#!/bin/sh

# delete prev image
if [ -z "$(docker images -q messages:latest 2> /dev/null)" ]; then
	docker image rm -f messages
fi

docker image build --tag messages --file Dockerfile . 
