#!/bin/bash
#
# A bash script containing utility functions for reusability
#

# Checks of Docker is started if not starts it and returns after start is successful
checkAndStartDockerDaemon() {
  # https://medium.com/@valkyrie_be/quicktip-a-universal-way-to-check-if-docker-is-running-ffa6567f8426
  #Open Docker, only if is not running
  if (! docker stats --no-stream >/dev/null); then
    echo "Docker daemon not started"
    # On Mac OS this would be the terminal command to launch Docker
    open /Applications/Docker.app


    #Wait until Docker daemon is running and has completed initialisation
    echo "Waiting for Docker daemon to launch..."
    while (! docker stats --no-stream >/dev/null); do
      # Docker takes a few seconds to initialize
      sleep 1
    done
    echo "Docker daemon has started"
  fi
}

# Checks if the specified docker image exists in local docker repo
# the docker image name is passed as a param
# returns 0 for true and 1 for false
existsInLocal() {
  if [[ "$(docker images -q "$1" 2> /dev/null)" != "" ]]; then
    # echo "Docker image $1 exists in local docker repository"
    return 0
  else
    # echo "Docker image $1 does not exist in local docker repository"
    return 1
  fi
}

# Checks if the specified docker image exists in GCR
# the docker image name is passed as 2 params (name and version)
# returns 0 for true and 1 for false
existsInGCR() {
  if (gcloud artifacts docker images list us-central1-docker.pkg.dev/eco-spirit-404410/bazelgo/$1 | grep $2 > /dev/null); then
    # echo "Docker image $1:$2 exists in gcr repository"
    return 0
  else
    # echo "Docker image $1:$2 does not exist in gcr repository"
    return 1
  fi
}

# Pushes the specified docker image to GCR
# the docker image name is passed as 2 params (name and version)
pushToGCR() {
  echo "Tagging Docker Image"
  docker tag $1:$2 us-central1-docker.pkg.dev/eco-spirit-404410/bazelgo/$1 #this line (without the version) will tag it as latest
  docker tag $1:$2 us-central1-docker.pkg.dev/eco-spirit-404410/bazelgo/$1:$2

  echo "Pushing the Image to Google Container"
  docker push us-central1-docker.pkg.dev/eco-spirit-404410/bazelgo/$1
  docker push us-central1-docker.pkg.dev/eco-spirit-404410/bazelgo/$1:$2
}