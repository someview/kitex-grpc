# kitex-grpc
kitex grpc plugin, seperate message definition  from serviceinfo, define template easily.
you can define importPaths for the proto file that contains  multi service

## example
```
go run main.go -c ./example/config.json
```

## usage
- install this by `go install github.com/somview/kitex-grpc`

- use this
```
kitex-grpc --help
kitex-grpc -c config.json
kitex-grpc -json 'jsonstr', eg:  
- unix-like shell kitex-grpc  -json='{"Protos":[{"FilePath":"aaa.proto","OutputPath":"../pb"}]}'
- windows powershell  kitex-grpc -json='{\"Protos\":[{\"FilePath\":\"aaa.proto\",\"OutputPath\":\"../pb\"}]}'
```