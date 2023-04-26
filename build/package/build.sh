#!/bin/sh

DOCKER_BUILDKIT=1 docker build -t admgr:0.1 -f build/package/Dockerfile . --no-cache