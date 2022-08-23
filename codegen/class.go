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

type Class struct {
	goID               string
	comment            string
	isDataType         bool
	parents            map[*Class]*parentOption
	Members            []*Property
	fieldDepthsChecked bool
	depthAdjustMax     int
}

func newClass() *Class {
	return &Class{parents: map[*Class]*parentOption{}}
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

func (c *Class) TypeName() string {
	return c.goID
}
func (c *Class) Expects(typeName string) bool {
	return c.TypeName() == typeName
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
	c.buildStruct(code)
	// c.buildDepthAdjuster(code)
	// code.Comment(strings.Join(c.getDupedAccesors(), ", ")).Line()
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

func (c *Class) buildStruct(code *jen.Statement) {
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
		code := jen.Id(p.GetFieldID()).Interface().Tag(p.jsonTag()).Comment(p.comment)
		fields.Add(code)
	}
	code.Type().Id(c.TypeName()).Struct(fields...).Line()
}
