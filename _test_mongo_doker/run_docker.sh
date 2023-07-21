#!/bin/bash

CONTAINER_NAME="tgi-mongodb"

mkdir -p data
# Check if the container is already running
if [[ "$(docker ps -q -f name=$CONTAINER_NAME)" ]]; then
  echo "Stopping and removing the existing container..."
  docker stop $CONTAINER_NAME
  docker rm $CONTAINER_NAME
fi

# Start a new container
docker run -d -p 27017:27017 -v ${PWD}/data:/data/db --name $CONTAINER_NAME mongo
