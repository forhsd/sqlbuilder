package main

import (
	"fmt"
	"strings"
	"text/template"

	"fyne.io/fyne/v2"
	"github.com/forhsd/sqlbuilder/common/facade"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// golang使用Fyne写一个根据模板库text/template中的{{.variable}}生成前端绑定组件,拿到所有用户选择组件值并将值渲染至模板对应变量.
// 功能流程逻辑:
// 1. 正则提取模板中所有变量(.START_TIME, .END_TIME, .PAYMENT_TYPE_NAME)作为需要渲染的变量组件
// 2. 变量可供选择组件类型(文本输入,数字输入,下拉框),用户选择类型和值后传入后端
// 3. 后端根据变量值动态渲染模板,输出字符串到控制台
func main() {
	// Sample template
	tmplStr := `
	{{ define "slave1" }}
	select
	    payment_type_name
	from "ksl_trade_887f2181-eb06-4d77-b914-7c37c884952c_db".ksl_transaction_record
	where
	    acl.chmod
	    and is_delete = 0
	    and state = 4
	    {{ if and .START_TIME .END_TIME -}}
			and business_day between {{.START_TIME}}::date and {{.END_TIME}}::date
	    {{end -}}

	    {{ if .PAYMENT_TYPE_NAME -}}
	        and payment_type_name = any( {{.PAYMENT_TYPE_NAME}}::text[])
	    {{end -}}

	
		{{if .PAYMENT_TYPE_NAME_A}}
			and payment_type_name_a = any( {{N1V .PAYMENT_TYPE_NAME_A}}::text[])
		{{end}}
	
	group by
	    payment_type_name
	order by
	    payment_type_name
	{{ end }}
	`

	// 匹配SQL是否存在模板: \{\{\s*define\s*"master"\s*\}\}[\s\S]*?\{\{\s*end\s*\}\}

	// 正则会误伤, 不建议使用
	// Extract variables using regex
	//re := regexp.MustCompile(`\{\{\.\s*(\w+)\s*\}\}`)
	//matches := re.FindAllStringSubmatch(tmplStr, -1)
	//
	//for _, match := range matches {
	//	vars[match[1]] = ""
	//}

	varNames, _ := facade.ExtraValFromTemplate(tmplStr)

	// 测试参数值: 2023-07-14 03:15:40
	vars := make(map[string]string)
	for _, v := range varNames {
		vars[v] = ""
	}

	// Create a new Fyne app
	myApp := app.New()
	myWindow := myApp.NewWindow("Template Renderer")

	// Create UI components for each variable
	entries := make(map[string]*widget.Entry)
	for v := range vars {
		entries[v] = widget.NewEntry()
	}

	// Layout
	formItems := make([]*widget.FormItem, 0)
	for v, entry := range entries {
		formItems = append(formItems, widget.NewFormItem(v, entry))
	}
	form := widget.NewForm(formItems...)

	// Render button
	renderButton := widget.NewButton("Render Template", func() {
		// Collect values from entries
		for v, entry := range entries {
			vars[v] = entry.Text
		}

		// Parse and execute the template
		t, err := template.New("query").Parse(tmplStr)
		if err != nil {
			fmt.Println("Error parsing template:", err)
			return
		}

		var renderedOutput strings.Builder
		err = t.ExecuteTemplate(&renderedOutput, "slave1", vars)
		if err != nil {
			fmt.Println("Error executing template:", err)
			return
		}

		// Output the result
		fmt.Println("Rendered Output:")
		fmt.Println(renderedOutput.String())
	})

	// Layout
	content := container.NewVBox(form, renderButton)
	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(400, 400))
	myWindow.ShowAndRun()
}
