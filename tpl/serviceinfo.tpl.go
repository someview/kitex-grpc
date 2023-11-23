package tpl

const ServiceInfoTpl = `

package {{.PackageName}}
{{ $outer := . }}

import (
	"context"
	kitex "github.com/cloudwego/kitex/pkg/serviceinfo"
	{{- range .ImportPaths}}
	. {{.}}
    {{- end}}
	streaming "github.com/cloudwego/kitex/pkg/streaming"
    "errors"
)

{{range .ServiceList}}
{{- $serviceName := .ServiceName }}

// {{$serviceName}} is designed to provide IDL-compatible methods with call-option parameter for kitex framework.
type {{$serviceName}} interface {
	{{- range .Methods}}
		{{- if or .ClientStreaming .ServerStreaming}}
	{{.Name}}(stream {{$serviceName}}_{{.Name}}Server) (err error)
		{{- else}}
	{{.Name}}(ctx context.Context, req *{{.RequestType}}) (resp *{{.ReturnType}}, err error)
		{{- end}}
	{{- end}}
}


    {{range .Methods}}
	{{- if or .ClientStreaming .ServerStreaming}}

type {{$serviceName}}_{{.Name}}Server interface {
	streaming.Stream
	{{- if .ClientStreaming}}
	Recv() (*{{.RequestType}}, error)
	{{- end}}
	{{- if .ServerStreaming}}
	Send(*{{.ReturnType}}) error
	{{- end}}
}
    {{- end}}
	{{- end}}


var {{.ServiceName}}ServiceInfo = New{{.ServiceName}}ServiceInfo()

func New{{.ServiceName}}ServiceInfo() *kitex.ServiceInfo {
	serviceName := "{{.ServiceName}}"
	handlerType := (*{{.ServiceName}})(nil)    
	methods := map[string]kitex.MethodInfo{
		{{- range .Methods}}
		    {{- if or .ClientStreaming .ServerStreaming}}
		"{{.Name}}": kitex.NewMethodInfo({{.HandlerName}}, func()any{return new({{.RequestType}})}, func()any{return new({{.ReturnType}})}, false),
		    {{- else}}
		"{{.Name}}": kitex.NewMethodInfo({{.HandlerName}}, func()any{return new({{.RequestType}})}, func()any{return new({{.ReturnType}})}, false),    
		    {{- end}}
		{{- end}}
	}
	extra := map[string]interface{}{
		"PackageName":     "{{$outer.PackageName}}",
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

// stream IDL	
type {{$serviceName}}{{.Name}}Server struct {
	streaming.Stream
}

{{- if .ClientStreaming}}

func (c *{{$serviceName}}{{.Name}}Server) Recv()(*{{.RequestType}}, error){
	m := new({{.RequestType}})
	return m, c.Stream.RecvMsg(m)
}
{{- end}}


{{- if .ServerStreaming}}

func (c *{{$serviceName}}{{.Name}}Server) Send(res *{{.ReturnType}}) (error) {
	return c.Stream.SendMsg(res)
}
{{- end}}


func {{.HandlerName}}(ctx context.Context, handler interface{}, arg, result interface{}) error {
	st := arg.(*streaming.Args).Stream
	stream := &{{$serviceName}}{{.Name}}Server{st}
	return handler.({{$serviceName}}).{{.Name}}(stream)
}
		{{- else}}
func {{.HandlerName}}(ctx context.Context, handler interface{}, arg, result interface{}) error {
    s, ok := arg.(*streaming.Args)
    if !ok {
       return  errors.New("{{$serviceName}} service must use grpc protocol")
    }
    
    st := s.Stream
    req := new({{.RequestType}})
    if err := st.RecvMsg(req); err != nil {
        return err
    }
    resp, err := handler.({{$serviceName}}).{{.Name}}(ctx, req)
    if err != nil {
        return err
    }
    if err := st.SendMsg(resp); err != nil {
        return err
    }
	return nil
}	
	{{- end}}
    {{- end}}

{{- end}}
`
