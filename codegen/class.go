package codegen

import (
	"sort"
	"strings"

	"github.com/dave/jennifer/jen"
)

type Property struct {
	Name    string
	Comment string
	Types   []string
}

type Class struct {
	RawID      string
	Comment    string
	isDataType bool
	Parents    []*Class
	Members    []*Property
}

func newClass() *Class {
	return &Class{}
}

func (c *Class) GoID() string {
	switch c.RawID {
	case "3DModel":
		return "ThreeDModel"
	}
	return c.RawID
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
		return strings.Compare(c.Parents[i].RawID, c.Parents[j].RawID) < 0
	})

	var fields = jen.Statement{}
	for i, p := range c.Parents {
		if i == 0 {
			fields.Add(jen.Id(p.GoID()))
		} else {
			fields.Add(jen.Comment("TODO: " + p.GoID()))
		}
	}
	for i, p := range c.Members {
		if i == 0 && len(fields) != 0 {
			fields.Line()
		}

		code := jen.Id(strings.Title(p.Name))

		if len(p.Types) == 1 {
			code.Id(p.Types[0] + "Values")
		} else {
			code.Id("Values")
		}

		code.Tag(map[string]string{"json": p.Name + ",omitempty"})
		code.Comment(p.Comment)

		fields.Add(code)
	}

	code := jen.Comment(c.Comment).Line()
	code.Type().Id(c.GoID()).Struct(fields...).Line()

	{ // Values
		code.Type().Id(c.GoID() + "Values").Index().Interface().Line()
		code.Func().Params(jen.Id("v").Op("*").Id(c.GoID() + "Values")).Id("UnmarshalJSON").Params(
			jen.Id("data").Index().Byte(),
		).Error().Block(
			jen.Return().Id("decodeObjectsAs").Call(
				jen.Parens(jen.Op("*").Index().Interface()).Call(jen.Id("v")),
				jen.Id("data"),
				jen.Lit(c.GoID()),
			),
		)
	}

	return code
}
