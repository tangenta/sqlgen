package sqlgen

func buildReplacer() *Replacer {
	r := &Replacer{}
	r.add("column_def", constStrFn("a int"))
	r.add("opt_temporary", constStrFn(""))
	r.add("bit_expr", constStrFn("3"))
	return r
}

func BuildContext(productions []*Production, replacer *Replacer) (ctx *Context) {
	return &Context{
		productionMap: buildProdMap(productions),
		replacer:      replacer,
	}
}

func constStrFn(str string) func() string {
	return func() string { return str }
}
