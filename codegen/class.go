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
	goID       string
	comment    string
	isDataType bool
	parents    []*Class
	Members    []*Property
}

func newClass() *Class {
	return &Class{}
}

func (c *Class) getParents() []*Class {
	parents := make([]*Class, len(c.parents))
	i := 0
	for _, p := range c.parents {
		parents[i] = p
		i++
	}
	sort.Slice(parents, func(i, j int) bool {
		return parents[i].TypeName() < parents[j].TypeName()
	})
	return parents
}

func (c *Class) TypeName() string {
	return c.goID
}

func (c *Class) IsDataType() bool {
	for _, p := range c.parents {
		if p.IsDataType() {
			return true
		}
	}
	return c.isDataType || c.TypeName() == "DataType"
}

func (c *Class) Code() jen.Code {
	sort.Slice(c.Members, func(i, j int) bool {
		return strings.Compare(c.Members[i].Name, c.Members[j].Name) < 0
	})

	var fields = jen.Statement{}
	for i, p := range c.getParents() {
		if i == 0 {
			fields.Add(jen.Id(p.TypeName()))
		} else {
			fields.Add(jen.Comment("TODO: " + p.TypeName()))
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

	code := jen.Comment(c.comment).Line()
	code.Type().Id(c.TypeName()).Struct(fields...).Line()
	return code
}
