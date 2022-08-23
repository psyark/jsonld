package codegen

import (
	"sort"
	"strings"

	"github.com/dave/jennifer/jen"
)

type parentOption struct {
	depthAdjust int
}

type providerMapEntry struct {
	Provider *Class
	Parent   *Class
	Depth    int
}
type providerMap map[string][]providerMapEntry

func (pmap providerMap) SortedKeys() []string {
	keys := make([]string, len(pmap))
	i := 0
	for k := range pmap {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}

type unionSet map[*Union]struct{}

func (us unionSet) Sorted() []*Union {
	sorted := make([]*Union, len(us))
	i := 0
	for u := range us {
		sorted[i] = u
		i++
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].TypeName() < sorted[j].TypeName()
	})
	return sorted
}

type Class struct {
	goID               string
	comment            string
	isDataType         bool
	parents            map[*Class]*parentOption
	Members            []*Property
	oneOf              unionSet
	fieldDepthsChecked bool
	depthAdjustMax     int
}

func newClass() *Class {
	return &Class{
		parents: map[*Class]*parentOption{},
		oneOf:   unionSet{},
	}
}

func (c *Class) getParents() []*Class {
	parents := make([]*Class, len(c.parents))
	i := 0
	for p := range c.parents {
		parents[i] = p
		i++
	}
	sort.Slice(parents, func(i, j int) bool {
		return parents[i].TypeName() < parents[j].TypeName()
	})
	return parents
}

func (c *Class) getAllUnions() unionSet {
	us := unionSet{}
	for k := range c.oneOf {
		us[k] = struct{}{}
	}
	for p := range c.parents {
		for k := range p.getAllUnions() {
			us[k] = struct{}{}
		}
	}
	return us
}

func (c *Class) TypeName() string {
	return c.goID
}
func (c *Class) Expects(typeName string) bool {
	return c.TypeName() == typeName
}
func (c *Class) StructName() string {
	s := c.goID
	return strings.ToLower(string([]rune(s)[0])) + s[1:] + "Struct"
}
func (c *Class) IsDataType() bool {
	for p := range c.parents {
		if p.IsDataType() {
			return true
		}
	}
	return c.isDataType || c.TypeName() == "DataType"
}

func (c *Class) Code() jen.Code {
	sort.Slice(c.Members, func(i, j int) bool {
		return strings.Compare(c.Members[i].goID, c.Members[j].goID) < 0
	})

	code := jen.Comment(c.comment).Line()
	c.buildInterface(code)
	c.buildStruct(code)
	c.buildDepthAdjuster(code)
	c.buildImplementMethods(code)
	c.buildReImplementMethods(code)
	c.buildAccessors(code)
	c.buildUnmarshal(code)
	BuildSlice(code, c)
	// code.Comment(strings.Join(c.getDupedAccesors(), ", ")).Line()
	// 実装確認
	code.Var().Id("_").Id(c.TypeName()).Op("=").Op("&").Id(c.StructName()).Block().Line()
	return code
}

// メンバーごとのプロバイダーを返す
func (c *Class) getMemberProviderMap() providerMap {
	pmap := providerMap{}
	for _, p := range c.Members {
		pmap[p.goID] = append(pmap[p.goID], providerMapEntry{c, nil, 0})
	}
	for _, p := range c.getParents() {
		po := c.parents[p]
		for k, entries := range p.getMemberProviderMap() {
			for _, e := range entries {
				pmap[k] = append(pmap[k], providerMapEntry{e.Provider, p, e.Depth + 1 + po.depthAdjust})
			}
		}
	}
	return pmap
}

func (c *Class) AdjustFieldDepths() {
	if !c.fieldDepthsChecked {
		// 親のチェックを先にやる
		for _, p := range c.getParents() {
			p.AdjustFieldDepths()
		}

		for p := c.getDupedFieldDepthParent(); p != nil; p = c.getDupedFieldDepthParent() {
			c.parents[p].depthAdjust++
			if p.depthAdjustMax < c.parents[p].depthAdjust {
				p.depthAdjustMax = c.parents[p].depthAdjust
			}
		}

		c.fieldDepthsChecked = true
	}
}

// 何らかのフィールドが同じ深さを持つParentを返す
func (c *Class) getDupedFieldDepthParent() *Class {
	pmap := c.getMemberProviderMap()
	for _, key := range pmap.SortedKeys() {
		m := map[int]bool{}
		for _, e := range pmap[key] {
			if m[e.Depth] {
				return e.Parent
			}
			m[e.Depth] = true
		}
	}
	return nil
}

func (c *Class) buildInterface(code *jen.Statement) {
	methods := jen.Statement{}

	if valueType := c.getValueType(); valueType != nil {
		code := jen.Id("Value").Call().Add(valueType)
		methods.Add(code)
	}

	for i, p := range c.getParents() { // アノニマスフィールド
		if i == 0 {
			methods.Comment("Parents")
		}
		methods.Add(jen.Id(p.TypeName()))
	}
	for i, u := range c.oneOf.Sorted() { // このクラスを代入可能なユニオン
		if i == 0 {
			methods.Comment("Unions")
		}
		methods.Add(jen.Id(u.TypeName()))
	}
	for i, p := range c.Members {
		if i == 0 {
			methods.Comment("Accessors")
		}
		methods.Add(jen.Id(p.GetMethodID()).Call().Index().Id(p.Type.TypeName()).Comment(p.comment))
	}
	code.Type().Id(c.TypeName()).Interface(methods...).Line()
}
func (c *Class) buildStruct(code *jen.Statement) {
	var fields = jen.Statement{}

	if valueType := c.getValueType(); valueType != nil {
		code := jen.Id("value").Add(valueType)
		fields.Add(code)
	}

	for _, p := range c.getParents() {
		code := jen.Id(p.StructName() + strings.Repeat("_", c.parents[p].depthAdjust))
		fields.Add(code)
	}
	for i, p := range c.Members {
		if i == 0 && len(fields) != 0 {
			fields.Line()
		}
		code := jen.Id(p.GetFieldID()).Id(SliceTypeName(p.Type)).Tag(p.jsonTag())
		fields.Add(code)
	}
	code.Type().Id(c.StructName()).Struct(fields...).Line()
}

func (c *Class) getReceiver() *jen.Statement {
	return jen.Id("s").Op("*").Id(c.StructName())
}

// 深さ調整
func (c *Class) buildDepthAdjuster(code *jen.Statement) {
	for i := 0; i < c.depthAdjustMax; i++ {
		code.Type().Id(c.StructName() + strings.Repeat("_", i+1)).Struct(
			jen.Id(c.StructName() + strings.Repeat("_", i)),
		).Line()
	}
}

// プロパティ実装メソッド
func (c *Class) buildImplementMethods(code *jen.Statement) {
	for _, u := range c.oneOf.Sorted() {
		code.Func().Call(c.getReceiver()).Id(u.SignatureMethodName()).Call().Block().Line()
	}
	if len(c.oneOf) != 0 {
		code.Line()
	}
}

// プロパティ再実装メソッド
// 複数の直接の親が同一のUnionを実装している場合、再実装する
func (c *Class) buildReImplementMethods(code *jen.Statement) {
	unionMap := map[*Union]int{}
	for _, p := range c.getParents() {
		for _, u := range p.getAllUnions().Sorted() {
			unionMap[u]++
		}
	}

	unions := []*Union{}
	for u, count := range unionMap {
		if count != 1 {
			unions = append(unions, u)
		}
	}
	sort.Slice(unions, func(i, j int) bool {
		return unions[i].TypeName() < unions[j].TypeName()
	})
	for _, u := range unions {
		code.Func().Call(c.getReceiver()).Id(u.SignatureMethodName()).Call().Block().Comment("re").Line()
	}
}

// アクセサ
func (c *Class) buildAccessors(code *jen.Statement) {
	if valueType := c.getValueType(); valueType != nil {
		code.Func().Call(c.getReceiver()).Id("Value").Call().Add(valueType).Block(
			jen.Return().Id("s").Dot("value"),
		).Line()
	}
	for _, p := range c.Members {
		code.Func().Call(c.getReceiver()).Id(p.GetMethodID()).Call().Index().Id(p.Type.TypeName()).Block(
			jen.Return().Id("s").Dot(p.GetFieldID()),
		).Line()
	}
}

func (c *Class) buildUnmarshal(code *jen.Statement) {
	if valueType := c.getValueType(); valueType != nil {
		code.Func().Call(c.getReceiver()).Id("UnmarshalJSON").Call(jen.Id("data").Index().Byte()).Error().Block(
			jen.Return().Qual("encoding/json", "Unmarshal").Call(
				jen.Id("data"),
				jen.Op("&").Id("s").Dot("value"),
			),
		).Line()
		code.Func().Call(c.getReceiver()).Id("MarshalJSON").Call().Params(jen.Index().Byte(), jen.Error()).Block(
			jen.Return().Qual("encoding/json", "Marshal").Call(
				jen.Id("s").Dot("value"),
			),
		).Line()
	}
}

func (c *Class) getValueType() jen.Code {
	if c.IsDataType() {
		switch c.TypeName() {
		case "Boolean":
			return jen.Bool()
		case "Number":
			return jen.Float64()
		case "Text", "Date", "DateTime", "Time":
			return jen.String()
		}
	}
	return nil
}
