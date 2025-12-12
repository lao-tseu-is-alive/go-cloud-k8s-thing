[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=lao-tseu-is-alive_go-cloud-k8s-thing&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=lao-tseu-is-alive_go-cloud-k8s-thing)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=lao-tseu-is-alive_go-cloud-k8s-thing&metric=reliability_rating)](https://sonarcloud.io/summary/new_code?id=lao-tseu-is-alive_go-cloud-k8s-thing)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=lao-tseu-is-alive_go-cloud-k8s-thing&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=lao-tseu-is-alive_go-cloud-k8s-thing) [![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=lao-tseu-is-alive_go-cloud-k8s-thing&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=lao-tseu-is-alive_go-cloud-k8s-thing)
[![test](https://github.com/lao-tseu-is-alive/go-cloud-k8s-thing/actions/workflows/test.yml/badge.svg)](https://github.com/lao-tseu-is-alive/go-cloud-k8s-thing/actions/workflows/test.yml)
[![cve-trivy-scan](https://github.com/lao-tseu-is-alive/go-cloud-k8s-thing/actions/workflows/cve-trivy-scan.yml/badge.svg)](https://github.com/lao-tseu-is-alive/go-cloud-k8s-thing/actions/workflows/cve-trivy-scan.yml)
[![codecov](https://codecov.io/gh/lao-tseu-is-alive/go-cloud-k8s-thing/branch/main/graph/badge.svg?token=02AHW79CES)](https://codecov.io/gh/lao-tseu-is-alive/go-cloud-k8s-thing)
# go-cloud-k8s-thing
go-cloud-k8s-thing  is a Thing  microservice written in Golang using [JWT](https://jwt.io) authentication from  go-cloud-k8s-user-group.

_it showcases how to build a container image with nerdctl, in a secured way (scan of CVE done with Trivy) and how to deploy it on Kubernetes_

Scan of security issues and other vulnerabilities are done automatically **before** building a container image (using [Trivy](https://aquasecurity.github.io/trivy/))
inside a github action that triggers when a version tag is pushed to this repo. 

### Latest Docker Container Image
you can find all the versions of this 
[image and instructions to pull the images from the Packages section on the right part of this page](https://github.com/lao-tseu-is-alive/go-cloud-k8s-thing/pkgs/container/go-cloud-k8s-thing) 


### Requirements:

You can find functional and system [requirements](https://github.com/lao-tseu-is-alive/go-cloud-k8s-thing/blob/main/documentation/Requirements.md) used to design this service in the documentation folder.

### OpenApi:
We ensure the OpenAPI contract based approach is fulfilled by generating the Go server routes directly from the Yaml specification using [deepmap/oapi-codegen](https://github.com/deepmap/oapi-codegen)

The OpenApi 3.0 definition is available in [Yaml](https://raw.githubusercontent.com/lao-tseu-is-alive/go-cloud-k8s-thing/main/api/thing.yaml) and [JSON](https://raw.githubusercontent.com/lao-tseu-is-alive/go-cloud-k8s-thing/main/api/thing.json) format in the api directory. The nice and interactive [Swagger documentation](https://lao-tseu-is-alive.github.io/go-cloud-k8s-thing/) is available at this url  https://lao-tseu-is-alive.github.io/go-cloud-k8s-thing/



### Dependencies
[Echo: high performance, extensible, minimalist Go web framework](https://echo.labstack.com/)

[deepmap/oapi-codegen: OpenAPI Client and Server Code Generator](https://github.com/deepmap/oapi-codegen)

[pgx: PostgreSQL Driver and Toolkit](https://pkg.go.dev/github.com/jackc/pgx)

[Json Web Token for Go (RFC 7519)](https://github.com/cristalhq/jwt)


### Project Layout and conventions
This project uses the Standard Go Project Layout : https://github.com/golang-standards/project-layout

