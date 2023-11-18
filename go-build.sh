#!/bin/bash

cd app
if [[ -d "build" ]]; then
    rm -r build
fi
archs=(amd64 arm64 arm ppc64le ppc64 s390x) 
for arch in ${archs[@]} 
do
    env CGO_ENABLED=0 GOOS=linux GOARCH=${arch} go build -o ./build/qbrs_${arch} -v main.go
done 

echo "已编译以下平台："
echo "$(file ./build/qbrs_*)"