package xml

import (
	"fmt"
	"github.com/beevik/etree"
	"strings"
	"testing"
)

// TestPrintDuplicatePomDependency 可以打印出pom.xml中dependencyManagement定义的有重复的依赖
func TestPrintDuplicatePomDependency(t *testing.T) {
	path := "pom.xml"

	doc := etree.NewDocument()
	if err := doc.ReadFromFile(path); err != nil {
		panic(err)
	}

	propMp := make(map[string][]*PropRowNo)
	properties := doc.FindElement("./project/properties")
	for i, v := range properties.ChildElements() {
		propWithRowNo := new(PropRowNo)
		arr := propMp[v.Tag]
		if arr == nil {
			arr = make([]*PropRowNo, 0)
		}
		arr = append(arr, propWithRowNo)
		propMp[v.Tag] = arr

		propWithRowNo.Row = i
		propWithRowNo.Value = v.Text()
	}

	dependencyDefMp := make(map[string]*DependencyDef)
	dependencies := doc.FindElements("./project/dependencyManagement/dependencies/dependency")
	for row, v := range dependencies {
		groupId := v.FindElement("groupId").Text()
		artifactId := v.FindElement("artifactId").Text()

		mavenId := groupId + "::" + artifactId
		def := dependencyDefMp[mavenId]
		if def == nil {
			def = NewDependencyDef()
			dependencyDefMp[mavenId] = def
		}
		def.rows = append(def.rows, row)

		versionPlaceHolder := v.FindElement("version").Text()
		if !strings.Contains(versionPlaceHolder, "${") {
			fmt.Println(mavenId + " 版本" + versionPlaceHolder + "不是占位符，忽略")
			continue
		}
		versionPlaceHolder = versionPlaceHolder[2 : len(versionPlaceHolder)-1]
		propRowNos := propMp[versionPlaceHolder]
		if len(propRowNos) > 0 {
			def.propRows = append(def.propRows, propRowNos...)
		}
	}

	for k, v := range dependencyDefMp {
		flag := false
		if len(v.rows) > 1 {
			fmt.Printf("%s 存在重复定义，序号：%v\n", k, v.rows)
			flag = true
		}
		if len(v.propRows) > 1 {
			fmt.Printf("%s 存在多个版本定义，序号明细如下：\n", k)
			for _, propRow := range v.propRows {
				fmt.Printf("[%d]%s\n", propRow.Row, propRow.Value)
			}
			flag = true
		}
		if flag {
			fmt.Println()
		}
	}
}

type PropRowNo struct {
	Row       int
	Value     string
	PlainText string
}

type DependencyDef struct {
	groupId    string
	artifactId string
	rows       []int
	propRows   []*PropRowNo
	PlainText  string
}

func NewDependencyDef() *DependencyDef {
	result := new(DependencyDef)
	result.rows = make([]int, 0)
	result.propRows = make([]*PropRowNo, 0)
	return result
}
