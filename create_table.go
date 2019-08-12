package sqlgen

import (
	"fmt"
	"github.com/pingcap/log"
	"math/rand"
	"strings"
	"time"
)

type Context struct {
	productionMap map[string]*Production
	tidbParser    *Parser
	randConfig    *RandConfig
}

type RandConfig struct {
	replacer *Replacer
}

type Replacer struct {
	s map[string]func() string
}

func (r *Replacer) add(str string, strSupplier func() string) {
	if r.s == nil {
		r.s = make(map[string]func() string)
	}
	r.s[str] = strSupplier
}
func (r *Replacer) contains(str string) bool {
	_, ok := r.s[str]
	return ok
}
func (r *Replacer) run(str string) string {
	fn, ok := r.s[str]
	if !ok {
		panic(fmt.Sprintf("%s not found in replacer", str))
	}
	if fn == nil {
		panic(fmt.Sprintf("fn is nil for %s", str))
	}
	return fn()
}

func buildContext(productions []*Production) (ctx *Context) {
	ctx = &Context{
		productionMap: buildProdMap(productions),
		tidbParser:    NewParser(),
		randConfig: &RandConfig{
			replacer: &Replacer{},
		},
	}
	ctx.randConfig.replacer.add("column_def", constStrFn("a int"))
	ctx.randConfig.replacer.add("opt_temporary", constStrFn(""))
	ctx.randConfig.replacer.add("bit_expr", constStrFn("3"))
	//ctx.randConfig.replacer.add("opt_temporary", func()string{ return randomSQLStr("%empty", nil, ctx)[0]})
	return
}

func constStrFn(str string) func() string {
	return func() string { return str }
}


func RandomSQLStr(productionName string, ctx *Context) []string {
	counter := make(map[string]int)
	defer func() {
		for key, count := range counter {
			ctx.productionMap[key].maxLoop += count
		}
	}()
	return randomSQLStr(productionName, ctx, counter)
}

var nothing []string

func randomSQLStr(productionName string, ctx *Context, counter map[string]int) []string {
	cfg := ctx.randConfig
	if cfg.replacer.contains(productionName) {
		return []string{cfg.replacer.run(productionName)}
	}

	production, exist := ctx.productionMap[productionName]
	if !exist {
		panic(fmt.Sprintf("Production '%s' not found", productionName))
	}
	production.maxLoop -= 1
	counter[productionName] += 1

	seqs := filterMaxLoopAndZeroChance(production.bodyList, ctx.productionMap)
	if len(seqs) == 0 {
		log.Debug("exiting from " + productionName)
		return nothing
	}
	seqs = randomize(seqs)
	for _, s := range seqs {
		var sql []string
		containsException := false
		for _, item := range s.seq {
			if literalStr, isLiteral := literal(item); isLiteral {
				if literalStr != "" {
					sql = append(sql, literalStr)
				}
			} else {
				fragment := RandomSQLStr(item, ctx)
				if len(fragment) == 0 {
					containsException = true
					break
				} else {
					sql = append(sql, fragment...)
				}
			}
		}
		if !containsException {
			return sql
		}
	}
	return nothing
}

func randomize(old BodyList) BodyList {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(old), func(i, j int) { old[i], old[j] = old[j], old[i] })
	return old
}

func filterMaxLoopAndZeroChance(bodyList BodyList, prodMap map[string]*Production) BodyList {
	var ret BodyList
	for _, v := range bodyList {
		if v.randomFactor == 0 {
			continue
		}
		containsExceed := false
		for _, i := range v.seq {
			if _, isLit := literal(i); !isLit {
				prod, ok := prodMap[i]
				if !ok {
					panic(fmt.Sprintf("production %s not found", i))
				}
				if prod.maxLoop <= 0 {
					containsExceed = true
					break
				}
			}
		}
		if !containsExceed {
			ret = append(ret, v)
		}
	}
	return ret
}

func literal(token string) (string, bool) {
	if strings.HasPrefix(token, "'") && strings.HasSuffix(token, "'") {
		return strings.Trim(token, "'"), true
	}
	return "", false
}
