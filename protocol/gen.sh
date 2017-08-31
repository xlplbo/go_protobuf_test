#!/bin/bash

echo "start compile proto file ..."
echo "exec protoc --plugin protoc-gen-go(x64 elf)"

chmod +x ./protoc/bin/protoc
chmod +x ./protoc/bin/protoc-gen-go
./protoc/bin/protoc --plugin=protoc-gen-go=./protoc/bin/protoc-gen-go --go_out . *.proto


if [ $? -eq 0 ]; then
    echo "compile success!"
else
    echo "compile failed!"
fi