[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=lao-tseu-is-alive_go-cloud-k8s-thing&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=lao-tseu-is-alive_go-cloud-k8s-thing)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=lao-tseu-is-alive_go-cloud-k8s-thing&metric=reliability_rating)](https://sonarcloud.io/summary/new_code?id=lao-tseu-is-alive_go-cloud-k8s-thing)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=lao-tseu-is-alive_go-cloud-k8s-thing&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=lao-tseu-is-alive_go-cloud-k8s-thing) [![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=lao-tseu-is-alive_go-cloud-k8s-thing&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=lao-tseu-is-alive_go-cloud-k8s-thing)
[![test](https://github.com/lao-tseu-is-alive/go-cloud-k8s-thing/actions/workflows/test.yml/badge.svg)](https://github.com/lao-tseu-is-alive/go-cloud-k8s-thing/actions/workflows/test.yml)
[![cve-trivy-scan](https://github.com/lao-tseu-is-alive/go-cloud-k8s-thing/actions/workflows/cve-trivy-scan.yml/badge.svg)](https://github.com/lao-tseu-is-alive/go-cloud-k8s-thing/actions/workflows/cve-trivy-scan.yml)
[![codecov](https://codecov.io/gh/lao-tseu-is-alive/go-cloud-k8s-thing/branch/main/graph/badge.svg?token=02AHW79CES)](https://codecov.io/gh/lao-tseu-is-alive/go-cloud-k8s-thing)
# go-cloud-k8s-thing
go-cloud-k8s-thing  is a an thing  microservice written in Golang using JWT authentication from  go-cloud-k8s-user-group. 

_This project showcases how to build a container image with nerdctl, in a secured way (scan of CVE done with Trivy) and how to deploy it on Kubernetes_


## Dependencies
[Echo: high performance, extensible, minimalist Go web framework](https://echo.labstack.com/)

[deepmap/oapi-codegen: OpenAPI Client and Server Code Generator](https://github.com/deepmap/oapi-codegen)

[pgx: PostgreSQL Driver and Toolkit](https://pkg.go.dev/github.com/jackc/pgx)

[Json Web Token for Go (RFC 7519)](https://github.com/cristalhq/jwt)


## Project Layout and conventions
This project uses the Standard Go Project Layout : https://github.com/golang-standards/project-layout
