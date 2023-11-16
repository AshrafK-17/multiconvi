#!/bin/bash

SRC_ROOT_DIR="."
BLD_DIR="build"
WD=`pwd`
EXEC_NAME="goprogram"
IMAGE_NAME="ut-convertimage"
VERSION="1.0"

# Include contents of Utils.sh (which contains reusable functions)
[ -f ../utils.sh ] && . ../utils.sh


build() {
  echo ""
  echo "-----------------------------------------------------------------------"
  echo "Building ConvertImage Utility version: $VERSION"
  echo "-----------------------------------------------------------------------"
  echo ""

  if [ -d ${SRC_ROOT_DIR}/${BLD_DIR}/ ]; then
    echo "Removing temp build directory if it exists"
    rm -rf "${SRC_ROOT_DIR}/${BLD_DIR}/"
  fi

  echo "Creating a temp build directory"
  mkdir -p "${SRC_ROOT_DIR}/${BLD_DIR}/"
  echo "Copying from SRC_ROOT_DIR/gosrc to temp build directory"
  cp -r "${SRC_ROOT_DIR}/gosrc/." "${SRC_ROOT_DIR}/${BLD_DIR}/"

  export GOPATH="$WD"
  cd "${SRC_ROOT_DIR}/${BLD_DIR}/"
  GOOS=linux GOARCH=amd64 go build -o ${EXEC_NAME} || exit 1
  cd "$WD"

  if [ ! -f "${SRC_ROOT_DIR}/${BLD_DIR}/${EXEC_NAME}" ]; then
    echo "Error! Build was not successful!"
    exit 1
  fi

  echo "Building Docker Image"
  docker build -t $IMAGE_NAME:$VERSION .
}

# main routine
{
  while getopts ":fp" opt; do
    case $opt in
      f)
        FORCEBUILD=1
        ;;
      p)
        PUSH=1
        ;;
    esac
  done

  VERSION="$(git describe --tags)"

  IMAGE_FULLNAME=$IMAGE_NAME:$VERSION

  checkAndStartDockerDaemon

  if (existsInLocal $IMAGE_FULLNAME); then
    if [ "$FORCEBUILD" ]; then
      echo "Docker image $IMAGE_FULLNAME exists in local docker repository, Forcing a rebuild"
      build
      echo ""
    else
      echo "Docker image $IMAGE_FULLNAME exists in local docker repository, Ignoring rebuild"
    fi
  else
    echo "Docker image $IMAGE_FULLNAME does not exist in local docker repository, Building image"
    build
    echo ""
  fi

  if [ "$PUSH" ]; then
    echo "Pushing Docker image $IMAGE_FULLNAME to GCR Docker Repo"
    pushToGCR $IMAGE_NAME $VERSION
    echo ""
  fi

  echo "-----------------------------------------------------------------------"
  echo ""
}