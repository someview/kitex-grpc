package main

import (
	"log"
	"os"
	"strings"
	"text/template"

	"kitex-grpc/tpl"

	"github.com/jhump/protoreflect/desc/protoparse"
)

type ServiceInfo struct {
	ServiceName    string       // 服务名
	Methods        []MethodInfo // 服务下的方法列表
	ContainsStream bool         // 服务是否包含stream
}

type MethodInfo struct {
	Name            string // 方法名
	HandlerName     string // 提供给况下使用的handlerName
	RequestType     string // 请求类型
	ReturnType      string // 返回类型
	ClientStreaming bool
	ServerStreaming bool
}

type FileServiceInfo struct {
	PackageName string
	HasStream   bool // 文件是否包含stream
	ServiceList []ServiceInfo
}

func main() {
	parser := protoparse.Parser{
		ImportPaths: []string{"."}, // 更改为您的 .proto 文件所在的目录
	}
	fileDescs, err := parser.ParseFiles("aaa.proto") // 更改为您的 .proto 文件名
	if err != nil {
		log.Fatalf("Failed to parse proto file: %v", err)
	}

	fd := fileDescs[0]
	fileServiceInfo := FileServiceInfo{
		PackageName: extractPackageName(fd.GetFileOptions().GetGoPackage()),
	}

	for _, svc := range fd.GetServices() {
		serviceInfo := ServiceInfo{
			ServiceName: svc.GetName(),
		}
		for _, method := range svc.GetMethods() {
			if method.IsClientStreaming() || method.IsServerStreaming() {
				fileServiceInfo.HasStream = true
				serviceInfo.ContainsStream = true
			}
			serviceInfo.Methods = append(serviceInfo.Methods, MethodInfo{
				Name:            method.GetName(),
				HandlerName:     method.GetName() + "Handler", // 统一提供，避免在模板里面多次生成HandlerName，易于维护
				RequestType:     method.GetInputType().GetName(),
				ReturnType:      method.GetOutputType().GetName(),
				ClientStreaming: method.IsClientStreaming(),
				ServerStreaming: method.IsServerStreaming(),
			})
		}
		fileServiceInfo.ServiceList = append(fileServiceInfo.ServiceList, serviceInfo)
	}

	generateServiceInfoCode(fileServiceInfo)
	generateServerCode(fileServiceInfo)
	generateClientCode(fileServiceInfo)
}

func extractPackageName(goPackage string) string {
	parts := strings.Split(goPackage, ";")
	var packageName string
	if len(parts) > 1 {
		packageName = parts[1] // 如果存在分号，取分号后面的部分
	} else {
		packageName = goPackage // 否则返回整个字符串
	}
	return strings.Replace(packageName, ".", "_", -1) // 替换所有的点为下划线
}

// ServiceInfo 和 MethodInfo 的定义 ...
// generateClientCode 函数的定义 ...

// 模板字符串

// generateClientCode 使用模板生成代码
func generateClientCode(fileServiceInfo FileServiceInfo) {
	tmpl, err := template.New("client").Funcs(template.FuncMap{
		"ToLower": strings.ToLower,
	}).Parse(tpl.ClientTpl)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	file, err := os.Create("../pb/client.go")
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	err = tmpl.Execute(file, fileServiceInfo)
	if err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}
}

func generateServiceInfoCode(fileServiceInfo FileServiceInfo) {
	tmpl, err := template.New("service").Funcs(template.FuncMap{
		"ToLower": strings.ToLower,
	}).Parse(tpl.ServiceInfoTpl)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	file, err := os.Create("../pb/service.go")
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	err = tmpl.Execute(file, fileServiceInfo)
	if err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}
}

func generateServerCode(fileServiceInfo FileServiceInfo) {
	tmpl, err := template.New("server").Funcs(template.FuncMap{
		"ToLower": strings.ToLower,
	}).Parse(tpl.ServerTpl)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	file, err := os.Create("../pb/server.go")
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	err = tmpl.Execute(file, fileServiceInfo)
	if err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}
}
