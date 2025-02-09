#!/usr/bin/env bash

rm -rf ../openapi/*.gen.go
oapi-codegen -config oapi-conf.yaml -package openapi api-swagger.yaml > ../openapi/openapi_generated.gen.go

cp api-swagger.yaml ../ui/public/api-swagger.yaml

echo Typescript now
rm -rf tsout
docker run --rm -u "$(id -u):$(id -g)" -v "${PWD}:/local" openapitools/openapi-generator-cli:v7.4.0 generate -i /local/api-swagger.yaml -g typescript-fetch -o /local/tsout
chmod -R g+rw .

rm -rf ../ui/src/openapi
mkdir ../ui/src/openapi
cp -r tsout/* ../ui/src/openapi

rm -rf tsout
