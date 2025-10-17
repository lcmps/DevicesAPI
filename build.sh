#!/bin/bash

docker build -t devices-api -f Dockerfile .
docker compose up