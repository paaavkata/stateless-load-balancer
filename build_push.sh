#! /bin/bash

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

go build -o stateless-load-balancer-amd64 .

# Push to S3
aws s3 cp stateless-load-balancer-amd64 s3://artifactory-data-store/stateless-load-balancer-amd64 --acl private

export GOARCH=arm64

go build -o stateless-load-balancer-arm64 .

# Push to S3
aws s3 cp stateless-load-balancer-arm64 s3://artifactory-data-store/stateless-load-balancer-arm64 --acl private