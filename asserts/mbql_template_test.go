package facade

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/forhsd/sqlbuilder/common/facade"
	"github.com/stretchr/testify/assert"
)

func TestTemplateString(t *testing.T) {
	// 所有前端输入框输入值
	// 1. 对于 string类型的值, 需要使用单引号包裹: `'%s'`
	// 2. 数值类型直接使用: `%v`
	// 3. 其他类型, : `'%v'`
	data := map[string]any{
		"STORE_GUID":        `'ffd10705-4449-4439-9d81-84bb08600e3c'`,
		"GROUP_STORE_GUID":  `'4ee5f568-b096-44c6-8b58-4147ef36a3fb'`,
		"START_TIME":        `'2024-07-05 17:17:37'`,
		"END_TIME":          `'2024-09-05 17:17:37'`,
		"PAYMENT_TYPE_NAME": `'交通银行'`,
		"Guids":             []any{"2203b12d-618a-4974-b09f-86ce3d35d26a", "362401e3-7876-4427-b467-8474a0efa03c"},
		"PrivilegesFlag":    `'-2'`,
	}

	// 1. 创建模板画布, 并注册所有自定义函数
	tmpl := template.New("root").Funcs(template.FuncMap{
		"PowerList": facade.PowerList,
	})

	// 2. 提取提前执行块
	// 将模板文件解析到模板画布中
	// ParseFiles 当前文件相对路径, 非项目根路径
	tmpl, err := tmpl.ParseFiles(filepath.Join("trans1.tmpl"))
	if err != nil {
		fmt.Println("Error parsing files:", err)
		return
	}

	// 解析模板文件slave1部分
	var salve1Buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&salve1Buf, "slave1", data)
	if err != nil {
		fmt.Println("Error executing template:", err)
	}

	fmt.Println("................slave1................")
	fmt.Println(salve1Buf.String())

	// 3. 参数注入提前执行块字符串, 命名: 子模板名Slave1 + Ctx, 或者 slaveX
	data["slave1"] = salve1Buf.String()

	// 4. 执行画布主模板
	var masterBuf bytes.Buffer
	err = tmpl.ExecuteTemplate(&masterBuf, "master", data)
	if err != nil {
		panic(err)
	}

	fmt.Println("................最终sql................")
	fmt.Println(masterBuf.String())
	assert.EqualValues(t, true, strings.Contains(masterBuf.String(), "ffd10705-4449-4439-9d81-84bb08600e3c"))
}

// 1. bi 调用接口X, 获取模板字符串是否存在 子模板个数和主模板个数(子模板内容可能一样也可能不一样)
// 2. bi判断: 如果 子模板个数=0, 调用 /sqlbuilder/proto, 直接执行主模板
// 3. 如果 X返回子模板个数>0,
// bi->sb: analyzeSubTmpl(data map[string]any, tmpl string) map[string]string, 结果kv: slaveN(模板变量名)->sql
// bi 执行子模板sql得到一个row[0][0]的字符串值(nil插入空字符串), 将字符串值新增至data入参中
// bi->sb: analyzeMasterTmpl(dataPush map[string]any, tmpl string) map[string]string, 结果kv: master->sql
// 4. bi->sb: /sqlbuilder/proto

func TestTemplateFlowWithoutNormalize(t *testing.T) {

	// 1. 模拟: 模板变量(所有有值)参数
	// 由于子模板(动态列)需要经由数据库执行才能得到完整值, 因此渲染子模板必须携带所有参数
	args := map[string]any{
		"STORE_GUID":        `'ffd10705-4449-4439-9d81-84bb08600e3c'`,
		"GROUP_STORE_GUID":  `'4ee5f568-b096-44c6-8b58-4147ef36a3fb'`,
		"START_TIME":        `'2024-07-05 17:17:37'`,
		"END_TIME":          `'2024-09-05 17:17:37'`,
		"PAYMENT_TYPE_NAME": `'交通银行'`,
		"Guids":             []any{"2203b12d-618a-4974-b09f-86ce3d35d26a", "362401e3-7876-4427-b467-8474a0efa03c"},
		"PrivilegesFlag":    `'-2'`,
	}

	// 2. 模拟: 读取整个模板字符串
	data, err := os.ReadFile(filepath.Join("trans1.tmpl"))
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// 将字节切片转换为字符串
	content := string(data)

	// 3. 解析提取子模板和主模板, 响应调用方, 调用方执行子模板得到结果后push参数到主模板data中
	// 没有嵌套时, len(templates.Slave) == 0
	templates, err := facade.AnalyzeSubTmpl(content, args)
	if err != nil {
		fmt.Println("解析模板失败:", err)
	}

	fmt.Println("子模板内容:")
	for k, v := range templates.Slave {
		fmt.Printf("模板名: %s \n", k)
		fmt.Printf("解析模板: %s", v)
	}

	fmt.Println("=============================分割线=============================")
	fmt.Println("主模板内容:")
	// 4. 封装 /sqlbuilder/proto 主模板查询请求
	fmt.Println(templates.Master)
	assert.EqualValues(t, false, strings.Contains(templates.Master, "ffd10705-4449-4439-9d81-84bb08600e3c"))
}

func TestTemplateFlowWithNormalize(t *testing.T) {

	// 1. 模拟: 模板变量(所有有值)参数
	// 由写SQL的用户来控制参数的规范化, 避免程序控制出问题
	args := map[string]any{
		"STORE_GUID":        `ffd10705-4449-4439-9d81-84bb08600e3c`,
		"GROUP_STORE_GUID":  `4ee5f568-b096-44c6-8b58-4147ef36a3fb`,
		"START_TIME":        `2024-07-05 17:17:37`,
		"END_TIME":          `2024-09-05 17:17:37`,
		"PAYMENT_TYPE_NAME": `交通银行`,
		"Guids":             []any{"2203b12d-618a-4974-b09f-86ce3d35d26a", "362401e3-7876-4427-b467-8474a0efa03c"},
		"PrivilegesFlag":    -2,
	}

	// 2. 模拟: 读取整个模板字符串
	data, err := os.ReadFile(filepath.Join("trans2.tmpl"))
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// 将字节切片转换为字符串
	content := string(data)

	// 3. 解析提取子模板和主模板, 响应调用方, 调用方执行子模板得到结果后push参数到主模板data中
	// 没有嵌套时, len(templates.Slave) == 0
	templates, err := facade.AnalyzeSubTmpl(content, args)
	if err != nil {
		fmt.Println("解析模板失败:", err)
	}

	fmt.Println("子模板内容:")
	for k, v := range templates.Slave {
		fmt.Printf("模板名: %s \n", k)
		fmt.Printf("解析模板: %s", v)
	}

	fmt.Println("=============================分割线=============================")
	fmt.Println("主模板内容:")
	// 4. 封装 /sqlbuilder/proto 主模板查询请求
	// 主模板保留字面量
	fmt.Println(templates.Master)
	assert.EqualValues(t, false, strings.Contains(templates.Master, "ffd10705-4449-4439-9d81-84bb08600e3c"))
}

func TestTemplateFlowWithDefaultValue(t *testing.T) {

	// 1. 模拟: 模板变量(所有有值)参数
	// 由写SQL的用户来控制参数的规范化, 避免程序控制出问题
	args := map[string]any{
		"STORE_GUID":        `ffd10705-4449-4439-9d81-84bb08600e3c`,
		"GROUP_STORE_GUID":  `4ee5f568-b096-44c6-8b58-4147ef36a3fb`,
		"START_TIME":        `2024-07-05 17:17:37`,
		"END_TIME":          `2024-09-05 17:17:37`,
		"PAYMENT_TYPE_NAME": `交通银行`,
		"Guids":             []any{"2203b12d-618a-4974-b09f-86ce3d35d26a", "362401e3-7876-4427-b467-8474a0efa03c"},
		"PrivilegesFlag":    -2,
		// "ENTERPRISE_ID":     `ffd10705-4449-4439-9d81-84bb08600e3c`,
	}

	// 1. 创建模板画布, 并注册所有自定义函数
	tmpl := template.New("root").Funcs(facade.InjectFunc())

	// 2. 提取提前执行块
	// 将模板文件解析到模板画布中
	// ParseFiles 当前文件相对路径, 非项目根路径
	tmpl, err := tmpl.ParseFiles(filepath.Join("trans3.tmpl"))
	if err != nil {
		fmt.Println("Error parsing files:", err)
		return
	}

	// 4. 执行画布主模板
	var masterBuf bytes.Buffer
	err = tmpl.ExecuteTemplate(&masterBuf, "master", args)
	if err != nil {
		panic(err)
	}

	fmt.Println("................最终sql................")
	fmt.Println(masterBuf.String())

	fmt.Println("=============================分割线1=============================")

}

func TestFormatBufAppend(t *testing.T) {
	var finalFormat string

	basicFormat := `avg(%s)`

	// 期望 flag1 = true, 在之前字符串追加小括号
	if true {
		finalFormat = fmt.Sprintf(`(%s)`, basicFormat)
	}
	// 期望 flag2 = true, 在之前字符串追加as
	if true {
		finalFormat = fmt.Sprintf(`%s as "%s"`, finalFormat, "别名")
	}

	if finalFormat == "" {
		finalFormat = basicFormat
	}

	// 打印最终 format 字符串
	fmt.Println(finalFormat)
}
