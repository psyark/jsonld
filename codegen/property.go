package build

import (
	"strings"

	"github.com/dave/jennifer/jen"
)

// メモ: rdf:Property は単に（クラスメンバーとしての）プロパティだけでなく、
// 型であり、代入されるデータの意味も定義している（っぽい）
// 例えば、「身長」と「体重」はどちらもfloatで表せるが
// rdf:Property では別のプロパティ（と不可分の型）が用意される。
// このため、とあるクラスに同じプロパティは重複して出現しない。

type Property struct {
	goID    string
	comment string
	Type    PropertyType
}

func (p *Property) GetImplementID() string {
	return "implements" + strings.Title(p.goID)
}
func (p *Property) GetFieldID() string {
	return strings.Title(p.goID) + "_"
}
func (p *Property) GetMethodID() string {
	return strings.Title(p.goID)
}
func (p *Property) jsonTag() map[string]string {
	return map[string]string{"json": p.goID + ",omitempty"}
}

type PropertyType interface {
	TypeName() string
	Expects(string) bool
}

var _ = []PropertyType{&Class{}, &Union{}}

func SliceTypeName(pt PropertyType) string {
	return pt.TypeName() + "Slice"
}

func BuildSlice(code *jen.Statement, pt PropertyType) {
	optionsDict := jen.Dict{}
	for _, typeName := range []string{"Text", "Date", "DateTime", "Time", "URL", "CssSelectorType", "XPathType"} {
		if pt.Expects(typeName) {
			optionsDict[jen.Id("Expect"+typeName)] = jen.Lit(true)
		}
	}
	options := jen.Id("DecodeOptions").Values(optionsDict)

	code.Type().Id(SliceTypeName(pt)).Index().Id(pt.TypeName()).Line()
	code.Func().Call(
		jen.Id("s").Op("*").Id(SliceTypeName(pt)),
	).Id("UnmarshalJSON").Call(jen.Id("data").Index().Id("byte")).Id("error").Block(
		// jen.Qual("fmt", "Println").Call(jen.Lit(SliceTypeName(pt)+".UnmarshalJSON"), jen.Id("string").Call(jen.Id("data"))),
		jen.Id("things").Op(",").Id("err").Op(":=").Id("DecodeObjects").Call(jen.Id("data"), options),
		jen.If(jen.Id("err").Op("!=").Nil()).Block(jen.Return().Id("err")),
		jen.Op("*").Id("s").Op("=").Id("make").Call(
			jen.Id(SliceTypeName(pt)).Op(",").Id("len").Call(jen.Id("things")),
		),
		jen.For(jen.Id("i").Op(",").Id("t").Op(":=").Range().Id("things")).Block(
			jen.Call(jen.Op("*").Id("s")).Index(jen.Id("i")).Op("=").Id("t").Op(".").Call(jen.Id(pt.TypeName())),
		),
		jen.Return().Nil(),
	).Line()
}
