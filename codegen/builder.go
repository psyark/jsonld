package codegen

import (
	"sort"
	"strings"

	"github.com/dave/jennifer/jen"
)

type Builder struct {
	classMap map[string]*Class
}

func NewBuilder(d Document) *Builder {
	b := &Builder{
		classMap: map[string]*Class{},
	}
	for _, g := range d.Graph {
		if g.HasType("rdfs:Class") {
			// クラス作成
			c := b.getClass(g.ID)
			c.goID = g.GetGoID()
			c.comment = g.Comment
			c.isDataType = g.HasType("schema:DataType")

			// 親クラスを設定
			for _, ref := range g.SubClassOf {
				if ref.ID != "rdfs:Class" {
					c.parents[b.getClass(ref.ID)] = &parentOption{}
				}
			}
		}
	}
	for _, g := range d.Graph {
		if g.HasType("rdf:Property") {
			p := &Property{
				Name:    strings.TrimPrefix(g.ID, "schema:"),
				goID:    g.GetGoID(),
				Comment: g.Comment,
			}

			if len(g.RangeIncludes) == 1 {
				// p.Type = b.getClass(g.RangeIncludes[0].ID)
			} else if len(g.RangeIncludes) > 1 {
				classes := []*Class{}
				for _, ref := range g.RangeIncludes {
					classes = append(classes, b.getClass(ref.ID))
				}

				// p.Type = union
				// p.Type = b.getClass(g.RangeIncludes[0].ID)
			}

			for _, ref := range g.DomainIncludes {
				c := b.getClass(ref.ID)
				// このプロパティをクラスメンバーに追加
				c.Members = append(c.Members, p)
			}
		}
	}

	return b
}

func (b *Builder) getClass(id string) *Class {
	if _, ok := b.classMap[id]; !ok {
		b.classMap[id] = newClass()
	}
	return b.classMap[id]
}

func (b *Builder) Build() error {
	if err := b.buildClasses(); err != nil {
		return err
	}

	return nil
}

func (b *Builder) buildClasses() error {
	classGo := jen.NewFile("jsonld")
	classGo.Comment("Code generated by jsonld.codegen; DO NOT EDIT.").Line()

	classNames := []string{}
	for k := range b.classMap {
		classNames = append(classNames, k)
	}
	sort.Strings(classNames)

	cases := jen.Statement{}
	for _, k := range classNames {
		c := b.classMap[k]
		if !c.IsDataType() {
			code := jen.Case(jen.Lit(c.TypeName())).Return().Op("&").Id(c.TypeName()).Block()
			cases.Add(code)
		}
	}
	classGo.Func().Id("NewThing").Call(jen.Id("name").String()).Interface().Block(
		jen.Switch(jen.Id("name")).Block(cases...),
		jen.Panic(jen.Id("name")),
	)

	for _, k := range classNames {
		classGo.Add(b.classMap[k].Code())
	}

	return classGo.Save("../gen_classes.go")
}
