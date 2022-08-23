package codegen

import (
	"sort"
	"strings"

	"github.com/dave/jennifer/jen"
)

type Property struct {
	Name    string
	Comment string
}

type Class struct {
	Name       string
	Comment    string
	isDataType bool
	Parents    []*Class
	Members    []*Property
}

func newClass() *Class {
	return &Class{}
}

func (c *Class) IsDataType() bool {
	for _, p := range c.Parents {
		if p.IsDataType() {
			return true
		}
	}
	return c.isDataType
}

func (c *Class) Code() jen.Code {
	sort.Slice(c.Members, func(i, j int) bool {
		return strings.Compare(c.Members[i].Name, c.Members[j].Name) < 0
	})
	sort.Slice(c.Parents, func(i, j int) bool {
		return strings.Compare(c.Parents[i].Name, c.Parents[j].Name) < 0
	})

	var fields = jen.Statement{}
	for i, p := range c.Parents {
		if i == 0 {
			fields.Add(jen.Id(p.Name))
		} else {
			fields.Add(jen.Comment("TODO: " + p.Name))
		}
	}
	for i, p := range c.Members {
		if i == 0 && len(fields) != 0 {
			fields.Line()
		}

		jsonTag := map[string]string{"json": p.Name + ",omitempty"}
		code := jen.Id(strings.Title(p.Name)).Id("Values").Tag(jsonTag).Comment(p.Comment)
		fields.Add(code)
	}

	code := jen.Comment(c.Comment).Line()
	code.Type().Id(c.Name).Struct(fields...).Line()
	return code
}
