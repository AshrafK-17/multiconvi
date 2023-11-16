#!/bin/bash

# A Bash utility script that helps start a single instance of ut-convertimage docker image

# Treat unset variables as a ERROR when executing bash commands
set -u

# Include contents of Utils.sh (which contains reusable functions)
[ -f ../utils.sh ] && . ../utils.sh


CONTAINER_NAME=dev-ut-convertimage
IMGNAME="ut-convertimage"

# main routine
{

  #VERSION="$(git describe --tags)"\
  VERSION="1.0"

  echo "Starting ut-convertimage docker container, using Image: $IMGNAME:$VERSION"
  docker run \
    -d \
    --privileged=true \
    --volume-driver=local \
    --name=$CONTAINER_NAME \
    --hostname=$CONTAINER_NAME \
    -p 9091:9091 \
    -e ENABLE_CORS=1 \
    -d $IMGNAME:$VERSION

  if [ $? -eq 0 ]; then
    echo "$CONTAINER_NAME successfully started"
    echo "To check logs:  docker logs $CONTAINER_NAME -f"
    echo "To stop container:  docker stop $CONTAINER_NAME && docker rm -v $CONTAINER_NAME"
  else
    echo "$CONTAINER_NAME failed to start"
  fi
}