# Windows environment

Install:
tdm-gcc
https://jmeubank.github.io/tdm-gcc/download/

# Publish image to docker repository

In .env, make sure that you set value for DOCKER_USERNAME and DOCKER_PASSWORD

```
yarn task build_linux
yarn task build_docker_image 0.01
yarn task push_docker_image 0.01
```

# CI

Run CI locally

```
cloud-build-local --config=build/cloudbuild-ci.yaml --dryrun=false ./
```