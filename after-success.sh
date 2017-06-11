#!/bin/bash
set -euo pipefail
IFS=$'\n\t'
REPO=quay.io/tfgco/maestro
DOCKER_REGISTRY=${DOCKER_REGISTRY:=quay.io}

$HOME/gopath/bin/goveralls -coverprofile _build/coverage-all.out -service=travis-ci

make build-docker

docker login -u $DOCKER_USER -p $DOCKER_PASSWORD $DOCKER_REGISTRY

# Logger image

# If this is not a pull request, update the branch's docker tag.
if [ $TRAVIS_PULL_REQUEST = 'false' ]; then
  docker tag maestro:latest $REPO:${TRAVIS_BRANCH/\//-} \
    && docker push $REPO:${TRAVIS_BRANCH/\//-};

  # If this commit has a tag, use on the registry too.
  if ! test -z $TRAVIS_TAG; then
    docker tag maestro:latest $REPO:${TRAVIS_TAG} \
      && docker push $REPO:${TRAVIS_TAG};
  fi
fi

REPO=quay.io/tfgco/maestro-config
# If this is not a pull request, update the branch's docker tag.
if [ $TRAVIS_PULL_REQUEST = 'false' ]; then
  docker tag maestro-config:latest $REPO:${TRAVIS_BRANCH/\//-} \
    && docker push $REPO:${TRAVIS_BRANCH/\//-};

  # If this commit has a tag, use on the registry too.
  if ! test -z $TRAVIS_TAG; then
    docker tag maestro-config:latest $REPO:${TRAVIS_TAG} \
      && docker push $REPO:${TRAVIS_TAG};
  fi
fi
