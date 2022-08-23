package codegen

import (
	"sort"
	"strings"

	"github.com/dave/jennifer/jen"
)

type Union struct {
	Classes []*Class
}

func (u *Union) TypeName() string {
	strs := []string{}
	for _, c := range u.Classes {
		strs = append(strs, c.TypeName())
	}
	sort.Strings(strs)
	return strings.Join(strs, "_") + "_Union"
}
func (u *Union) Expects(typeName string) bool {
	for _, c := range u.Classes {
		if c.Expects(typeName) {
			return true
		}
	}
	return false
}

func (u *Union) SignatureMethodName() string {
	return "implements" + u.TypeName()
}

func (u *Union) Code() jen.Code {
	// コメント
	code := jen.Comment(u.TypeName()).Line()

	// インターフェイス
	code.Type().Id(u.TypeName()).Interface(
		jen.Id(u.SignatureMethodName()).Call(),
	).Line()

	BuildSlice(code, u)

	// 実装確認
	body := []jen.Code{}
	for _, c := range u.Classes {
		code := jen.Id(c.TypeName()).Call(jen.Nil())
		body = append(body, code)
	}
	code.Var().Id("_").Op("=").Id(SliceTypeName(u)).Values(
		body...,
	).Line()

	return code
}
