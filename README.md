# go_protobuf_test
golang protobuf test

## Start
```
go get -u github.com/xlplbo/go_protobuf_test
```

## Introduce
How to compile proto file?

// window
```
./gen.bat
```
//linux64 
//if 32, download "protc" form https://github.com/google/protobuf/releases and compile plugin "protoc-gen-go"
```
$ sh gen.sh
```

How to demonstrate？
```
cd server
go build
./server
cd ../client
go build
./client
```

## Context
Use protobuf in golang.
