package tpl

var ServerTpl string = `
package {{.PackageName}}

import (
	server "github.com/cloudwego/kitex/server"
    
)
{{range .ServiceList}}
// NewServer creates a server.Server with the given handler and options.
func New{{.ServiceName}}Server(handler {{.ServiceName}}, opts ...server.Option) server.Server {
    var options []server.Option
    options = append(options, opts...)

    svr := server.NewServer(options...)
    if err := svr.RegisterService(New{{.ServiceName}}ServiceInfo(), handler); err != nil {
            panic(err)
    }
    return svr
}
{{- end}}
`
