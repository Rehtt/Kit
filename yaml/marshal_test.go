package yaml

import (
	"fmt"
	"testing"
)

func TestMarshalWithComment(t *testing.T) {
	type Test3 struct {
		B []string `yaml:"b" comment:"b"`
		Q string   `yaml:"q"`
		W int      `yaml:"w" comment:"WW"`
	}
	type Test2 struct {
	}
	type Test struct {
		A  string `yaml:"a" comment:"A"`
		T2 Test2  `yaml:"t2" comment:"t2"`
		T3 Test3  `yaml:"t3" comment:"t3"`
	}
	var tmp Test

	data, _ := MarshalWithComment(tmp)
	fmt.Println(string(data))
}

func TestName(t *testing.T) {
	type Test2 struct {
		C float64 `yaml:"c"`
	}
	type Test3 struct {
		B  []string   `yaml:"b" comment:"b"`
		Q  string     `yaml:"q"`
		W  int        `yaml:"w" comment:"WW"`
		T2 [][]*Test2 `yaml:"t2"`
	}
	type Test struct {
		A  string `yaml:"a" comment:"A"`
		T2 Test2  `yaml:"t2" comment:"t2"`
		T3 *Test3 `yaml:"t3" comment:"t3"`
	}
	var tmp Test
	data, _ := GenYamlTemplate(tmp)
	fmt.Println(string(data))
}
