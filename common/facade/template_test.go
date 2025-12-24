package facade

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forhsd/sqlbuilder/gen/proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
)

var protoJsonOP = protojson.MarshalOptions{
	EmitUnpopulated: true,
	// UseEnumNumbers:  true,
}

func TestAnalyzeTemplatesByProto(t *testing.T) {
	// 读取整个模板字符串
	data, err := os.ReadFile(filepath.Join("../../assets/trans2.tmpl"))
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// 将字节切片转换为字符串
	_ = string(data)

	req := &proto.AnalyzeTemplateRequest{
		Tmpl: "",
		Args: makeArgs(),
	}

	// protobuf 协议解析成json
	buf, err := protoJsonOP.Marshal(req)

	if err != nil {
		fmt.Println(err)
	}

	json := string(buf)
	fmt.Printf(json)
	assert.EqualValues(t, true, strings.Contains(json, "ffd10705-4449-4439-9d81-84bb08600e3c"))
}

func makeArgs() map[string]*proto.TemplateArg {
	args := make(map[string]*proto.TemplateArg)
	storeGuid := &proto.TemplateArg_StrVal{StrVal: "ffd10705-4449-4439-9d81-84bb08600e3c"}

	groupStoreGuid := &proto.TemplateArg_StrVal{StrVal: "4ee5f568-b096-44c6-8b58-4147ef36a3fb"}
	startTime := &proto.TemplateArg_StrVal{StrVal: "2024-07-05 17:17:37"}
	endTime := &proto.TemplateArg_StrVal{StrVal: "2024-09-05 17:17:37"}
	paymentTypeName := &proto.TemplateArg_StrVal{StrVal: "交通银行"}
	privilegesFlag := &proto.TemplateArg_IntVal{IntVal: -2}

	guids := makeArgValItems()

	args["STORE_GUID"] = &proto.TemplateArg{Data: storeGuid}
	args["GROUP_STORE_GUID"] = &proto.TemplateArg{Data: groupStoreGuid}
	args["START_TIME"] = &proto.TemplateArg{Data: startTime}
	args["END_TIME"] = &proto.TemplateArg{Data: endTime}
	args["PAYMENT_TYPE_NAME"] = &proto.TemplateArg{Data: paymentTypeName}
	args["Guids"] = &proto.TemplateArg{Data: guids}
	args["PrivilegesFlag"] = &proto.TemplateArg{Data: privilegesFlag}

	return args
}

func makeArgValItems() *proto.TemplateArg_ValItems {
	orgVals := make([]*proto.BasicData, 0)
	orgVals = append(orgVals, &proto.BasicData{Data: &proto.BasicData_StrVal{StrVal: "2203b12d-618a-4974-b09f-86ce3d35d26a"}})
	orgVals = append(orgVals, &proto.BasicData{Data: &proto.BasicData_StrVal{StrVal: "362401e3-7876-4427-b467-8474a0efa03c"}})
	arr := &proto.BasicDataArr{Args: orgVals}

	return &proto.TemplateArg_ValItems{ValItems: arr}
}

func TestGenSqlByTemplateProto(t *testing.T) {
	tmplSql := &proto.MixSql_Template{
		Template: &proto.SqlText{
			Text: "",
			// 还要加上 slaveN -> /sqlbuilder/template-analyze 解析后的 slave
			Args: makeArgs(),
		},
	}

	// 模板不管是否有提前查询块只有0层
	deep0 := &proto.DeepWrapper{Deep: 0, Sql: &proto.MixSql{Ref: tmplSql}}

	sbRequest := &proto.BuilderRequest{
		Driver:   proto.Driver_DRIVER_POSTGRES,
		Strategy: proto.BuilderStrategy_BUILDER_STRATEGY_TEMPLATE,
		Builders: []*proto.DeepWrapper{deep0},
	}

	buf, err := protoJsonOP.Marshal(sbRequest)

	if err != nil {
		fmt.Println(err)
	}

	json := string(buf)
	fmt.Printf(json)
	assert.EqualValues(t, true, strings.Contains(json, "ffd10705-4449-4439-9d81-84bb08600e3c"))
}
