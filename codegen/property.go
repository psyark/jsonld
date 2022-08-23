package codegen

import (
	"strings"
)

// メモ: rdf:Property は単に（クラスメンバーとしての）プロパティだけでなく、
// 型であり、代入されるデータの意味も定義している（っぽい）
// 例えば、「身長」と「体重」はどちらもfloatで表せるが
// rdf:Property では別のプロパティ（と不可分の型）が用意される。
// このため、とあるクラスに同じプロパティは重複して出現しない。

type Property struct {
	Name    string
	goID    string
	Comment string
}

func (p *Property) GetFieldID() string {
	return strings.Title(p.goID)
}

func (p *Property) jsonTag() map[string]string {
	return map[string]string{"json": p.Name + ",omitempty"}
}
