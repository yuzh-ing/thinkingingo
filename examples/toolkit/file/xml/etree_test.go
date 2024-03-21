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
		propRowNo := new(PropRowNo)
		arr := propMp[v.Tag]
		if arr == nil {
			arr = make([]*PropRowNo, 0)
		}
		arr = append(arr, propRowNo)
		propMp[v.Tag] = arr

		propRowNo.Row = i
		propRowNo.Value = v.Text()
		propRowNo.PlainText = fmt.Sprintf("<%s>%s<%s>", v.Tag, v.Text(), v.Tag)
	}

	dependencyDefMp := make(map[string]*DependencyDef)
	dependencies := doc.FindElements("./project/dependencyManagement/dependencies/dependency")
	for row, v := range dependencies {
		groupId := v.FindElement("groupId").Text()
		artifactId := v.FindElement("artifactId").Text()
		versionPlaceHolder := v.FindElement("version").Text()

		mavenId := groupId + "::" + artifactId
		def := dependencyDefMp[mavenId]
		if def == nil {
			def = NewDependencyDef()
			def.groupId = groupId
			def.artifactId = artifactId
			dependencyDefMp[mavenId] = def
		}
		def.dependencyRowNos = append(def.dependencyRowNos, &DependencyRowNo{
			row,
			groupId,
			artifactId,
			versionPlaceHolder,
		})

		if !strings.Contains(versionPlaceHolder, "${") {
			fmt.Println(mavenId + " 版本" + versionPlaceHolder + "不是占位符，忽略")
			continue
		}
		version := versionPlaceHolder[2 : len(versionPlaceHolder)-1]
		propRowNos := propMp[version]
		if len(propRowNos) > 0 {
			for _, propRowNo := range propRowNos {
				flag := true
				for _, defPropRow := range def.propRows {
					if defPropRow.Row == propRowNo.Row {
						flag = false
						break
					}
				}
				if flag {
					def.propRows = append(def.propRows, propRowNo)
				}
			}
		}
	}

	fmt.Println()
	for k, v := range dependencyDefMp {
		flag := false
		if len(v.dependencyRowNos) > 1 {
			fmt.Printf("%s 重复定义，明细如下：\n", k)
			for _, dependencyRow := range v.dependencyRowNos {
				fmt.Println("", dependencyRow.Text())
			}
			flag = true
		}
		if len(v.propRows) > 1 {
			fmt.Printf("%s 指定了多个版本，明细如下：\n", k)
			for _, propRow := range v.propRows {
				fmt.Println("  ", propRow.Text())
			}
			flag = true
		}
		if flag {
			fmt.Println("================================================================================")
		}
	}
}

type PropRowNo struct {
	Row       int
	Value     string
	PlainText string
}

func (pr *PropRowNo) Text() string {
	return pr.PlainText
}

type DependencyRowNo struct {
	Row        int
	groupId    string
	artifactId string
	version    string
}

func (dr *DependencyRowNo) Text() string {
	fmtStr := "<dependency>\n\t<groupId>%s</groupId>\n\t<artifactId>%s</artifactId>\n\t<version>%s</version>\n</dependency>"
	return fmt.Sprintf(fmtStr, dr.groupId, dr.artifactId, dr.version)
}

type DependencyDef struct {
	groupId          string
	artifactId       string
	dependencyRowNos []*DependencyRowNo
	propRows         []*PropRowNo
}

func NewDependencyDef() *DependencyDef {
	result := new(DependencyDef)
	result.dependencyRowNos = make([]*DependencyRowNo, 0)
	result.propRows = make([]*PropRowNo, 0)
	return result
}
