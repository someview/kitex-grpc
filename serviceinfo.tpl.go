package main

const serviceInfoTemplate = `

package {{.PackageName}}

import (
	"context"
	"fmt"
	client "github.com/cloudwego/kitex/client"
	kitex "github.com/cloudwego/kitex/pkg/serviceinfo"
	{{- if .HasStream}}
	"streaming" "github.com/cloudwego/kitex/pkg/streaming"
	{{- end}}
	proto "github.com/gogo/protobuf/proto"
)

{{range .ServiceList}}
{{- $serviceName := .ServiceName }}
func serviceInfo() *kitex.ServiceInfo {
	return {{.ServiceName}}ServiceInfo
}

var {{.ServiceName}}ServiceInfo = NewServiceInfo()

func New{{.ServiceName}}ServiceInfo() *kitex.ServiceInfo {
	serviceName := "{{.ServiceName}}"
	handlerType := ({{.ServiceName}})(nil)    
	methods := map[string]kitex.MethodInfo{
		{{- range .Methods}}
		"{{.Name}}": kitex.NewMethodInfo({{.HandlerName}}, &{{.RequestType}}, &&{{.ResponseType}}, false),
		{{- end}}
	}
	extra := map[string]interface{}{
		"PackageName":     "{{.PackageName}}",
		"ServiceFilePath": "",
	}
	{{- if .ContainsStream}}
	extra["streaming"] = true
	{{- end}}
	svcInfo := &kitex.ServiceInfo{
		ServiceName:     serviceName,
		HandlerType:     handlerType,
		Methods:         methods,
		PayloadCodec:    kitex.Protobuf,
		KiteXGenVersion: "v0.7.3",
		Extra:           extra,
	}
	return svcInfo
}

	{{- range .Methods}}
		{{- if or .ClientStreaming .ServerStreaming}}
type {{$serviceName}}{{.Name}}Server struct {
	streaming.Stream
}

func {{.HandlerName}}(ctx context.Context, handler interface{}, arg, result interface{}) error {
	st := arg.(*streaming.Args).Stream
	stream := &{{$serviceName}}{{.Name}}{st}
	return handler.({{$serviceName}}).{{.Name}}(stream)
}
		{{- else}}
func {{.HandlerName}}(ctx context.Context, handler interface{}, arg, result interface{}) error {
	switch s := arg.(type) {
	case *streaming.Args:
		st := s.Stream
		req := new({{.RequestType}})
		if err := st.RecvMsg(req); err != nil {
			return err
		}
		resp, err := handler.({{$serviceName}}).PostMessage(ctx, req)
		if err != nil {
			return err
		}
		if err := st.SendMsg(resp); err != nil {
			return err
		}
	case *{{.RequestType}}:
		res, err := handler.({{$serviceName}}).PostMessage(ctx, s.Req)
		if err != nil {
		    return err
		}
		*(result.({{**ResponseType}})) = res
	}
	return nil
}	
		{{- end}}	
    {{- end}}

{{- end}}
`
