#!/usr/bin/env bash

set -ex

readonly api_base="${IMAGE}_base:latest"

# Fetch base image, and invalidate it in case if anything has changed
docker pull "$api_base" || true

# Build base container
docker build -t "$api_base" --target base --cache-from "$api_base" .
docker push "$api_base"

# Build deployment container, using the base image as a cache
docker build \
       -t "${IMAGE}:$TAG" \
       --cache-from "$api_base" \
       --label "com.xing.git.sha1=${GIT_COMMIT}" \
       --label "com.xing.git.remote=${GIT_URL}" \
       --label com.xing.docker_build.target=release \
       .
