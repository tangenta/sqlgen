package sqlgen

import (
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func buildReplacer() *Replacer {
	r := &Replacer{}
	//r.add("column_def", constStrFn("a int"))
	r.add("opt_temporary", constStrFn(""))
	r.add("bit_expr", constStrFn("3"))
	r.add("expr", constStrFn("const_expr"))
	r.add("ident", generateIdent())
	r.add("opt_field_length", halfChance("(8)"))
	r.add("precision", halfChance("(3, 6)"))
	r.chainResetCallback(func() { r.add("ident", generateIdent()) })
	return r
}

func generateIdent() func() string {
	i := 0
	return func() string {
		i += 1
		return "ident" + strconv.Itoa(i)
	}
}

//type SQLType int
//const (
//	intType int = iota
//
//	maxNum
//)
//
//func generateType(r *Replacer) func() string {
//	return func() string {
//		randBranch := rand.Intn(maxNum)
//		switch randBranch {
//		case intType:
//			return randomSQL("int_type", r) +
//		}
//	}
//}

func BuildContext(productions []*Production, replacer *Replacer) (ctx *Context) {
	ctx = &Context{
		productionMap: BuildProdMap(productions),
		replacer:      replacer,
	}
	if replacer != nil {
		replacer.replacerCtx = ctx
	}
	return ctx
}

func constStrFn(str string) func() string {
	return func() string { return str }
}

func halfChance(str string) func() string {
	rand.Seed(time.Now().UnixNano())
	return func() string {
		chance := rand.Intn(2)
		if chance == 1 {
			return str
		}
		return ""
	}
}

func randomSQL(prodName string, replacer *Replacer) string {
	return strings.Join(RandomSQLStr(prodName, replacer.replacerCtx), " ")
}
