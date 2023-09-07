#!/bin/bash

# get_name() {
#     GOARCH=${1}
#     result=""
#         case "$GOARCH" in
#             amd64)
#                 result="x86_64"
#                 ;;
#             arm64)
#                 result="aarch64"
#                 ;;
#             arm)
#                 result="armv7l"
#                 ;;
#             ppc64le)
#                 result="ppc64le"
#                 ;;
#             ppc64)
#                 result="ppc64"
#                 ;;
#             s390x)
#                 result="s390x"
#                 ;;
#             *)
#                 result=""
#                 ;;
#         esac
#     echo "${result}"
# }
cd go
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