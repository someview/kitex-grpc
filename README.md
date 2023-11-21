# kitex-grpc
kitex grpc plugin, seperate message definition from serviceinfo, define template easily.
you can define importPaths for the proto file that contains multiservice.

## example
```
go run main.go -c ./example/config.json
```

## usage
- install this by `go install github.com/somview/kitex-grpc`

- use cases
```
kitex-grpc --help
kitex-grpc -c config.json
kitex-grpc -json 'jsonstr', eg:  
- unix-like shell kitex-grpc  -json='{"Protos":[{"FilePath":"aaa.proto","OutputPath":"../pb"}]}'
- windows powershell  kitex-grpc -json='{\"Protos\":[{\"FilePath\":\"aaa.proto\",\"OutputPath\":\"../pb\"}]}'
```

## examples
use with gogoproto
```
// message and service in the same module
protoc --gofast_out=. myproto.proto && kitex-grpc -json='{"Protos":[{"FilePath":"myproto.proto","OutputPath":"."}]}'
// message and service in the different module
protoc --gofast_out=. myproto.proto && kitex-grpc -json='{"Protos":[{"FilePath":"myproto.proto","OutputPath":"./serviceDirectory", "ImportPaths":["messageModulePath"...]}]}'
```
