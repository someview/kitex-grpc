package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/somview/kitex-grpc/tpl"
)

const protoSuffix = ".proto"

type Config struct {
	IncludePaths []string      `json:"IncludePaths"`
	Protos       []ProtoConfig `json:"Protos"`
}

type ProtoConfig struct {
	FilePath    string   `json:"FilePath"`    // proto文件名
	ImportPaths []string `json:"ImportPaths"` // 引入的msg的Module定义位置,允许引入多个module定义,例如github.com/someview/kitex-grpc
	OutputPath  string   `json:"OutputPath"`  // 指定的proto的输出路径
}

func parseConfig(configFile string, jsonConfig string) (*Config, error) {
	if jsonConfig != "" {
		var config Config
		err := json.Unmarshal([]byte(jsonConfig), &config)
		if err != nil {
			return nil, fmt.Errorf("error parsing JSON config: %v", err)
		}
		return &config, nil
	}

	if configFile != "" {
		configData, err := os.ReadFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("error reading config file: %v", err)
		}

		var config Config
		err = json.Unmarshal(configData, &config)
		if err != nil {
			return nil, fmt.Errorf("error parsing config file: %v", err)
		}
		return &config, nil
	}

	return nil, fmt.Errorf("no configuration provided")
}

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
	ProtoConfig
}

func main() {
	// 解析命令行参数或者配置文件参数
	configFile := flag.String("c", "", "Path to JSON config file,")
	help := `JSON string with configuration,{"IncludePaths":["path1","path2"],"Protos":[{"FilePath":"file1.proto","ImportPaths":["import1","import2"],"OutputPath":"output1"}]}"`
	jsonConfig := flag.String("json", "", help)
	flag.Parse()
	conf, err := parseConfig(*configFile, *jsonConfig)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	generateProtoFileSet(conf)
}

func generateProtoFileSet(conf *Config) {
	// importPath为默认路径,用于include
	incoludePaths := []string{"/", "."}
	incoludePaths = append(incoludePaths, conf.IncludePaths...)
	parser := &protoparse.Parser{
		ImportPaths:      incoludePaths,
		InferImportPaths: true,
	}
	for _, proto := range conf.Protos {
		fileServiceInfo := generateProtoFileInfo(parser, proto)
		if len(fileServiceInfo.ServiceList) > 0 {
			generateServiceInfoCode(fileServiceInfo)
			generateServerCode(fileServiceInfo)
			generateClientCode(fileServiceInfo)
		}
	}
}

func generateProtoFileInfo(parser *protoparse.Parser, conf ProtoConfig) FileServiceInfo {
	fmt.Println("conf.filePath:", conf.FilePath)
	fileDescs, err := parser.ParseFiles(conf.FilePath)
	if err != nil || len(fileDescs) == 0 {
		log.Fatalf("Failed to parse proto file: %v, err: %v", conf.FilePath, err)
	}

	fd := fileDescs[0]
	fileServiceInfo := FileServiceInfo{
		PackageName: extractPackageName(fd.GetFileOptions().GetGoPackage()),
		ProtoConfig: conf,
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
	return fileServiceInfo
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

func extractFileName(goProto string) string {
	parts := strings.Split(goProto, "/")
	var protoName string
	if len(parts) > 0 {
		protoName = strings.TrimSuffix(parts[len(parts)-1], protoSuffix)
	} else {
		protoName = strings.TrimSuffix(goProto, protoSuffix)
	}
	return protoName
}

// generateClientCode 使用模板生成代码
func generateClientCode(info FileServiceInfo) {
	tmpl, err := template.New("client").Parse(tpl.ClientTpl)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}
	filePath := fmt.Sprintf("%s/%s_client.go", info.OutputPath, extractFileName(info.FilePath))
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	err = tmpl.Execute(file, info)
	if err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}
}

func generateServiceInfoCode(info FileServiceInfo) {
	tmpl, err := template.New("service").Parse(tpl.ServiceInfoTpl)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}
	filePath := fmt.Sprintf("%s/%s_serviceinfo.go", info.OutputPath, extractFileName(info.FilePath))
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	err = tmpl.Execute(file, info)
	if err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}
}

func generateServerCode(info FileServiceInfo) {
	tmpl, err := template.New("server").Parse(tpl.ServerTpl)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}
	filePath := fmt.Sprintf("%s/%s_server.go", info.OutputPath, extractFileName(info.FilePath))
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	err = tmpl.Execute(file, info)
	if err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}
}
