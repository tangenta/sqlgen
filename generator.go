package sqlgen

import (
	"fmt"
	"github.com/pingcap/log"
	"math/rand"
	"strings"
)

var debugMode = false

type Context struct {
	productionMap map[string]*Production
	replacer      *Replacer
	trace         []string
}

type Replacer struct {
	replacerCtx   *Context
	s             map[string]func() string
	resetCallback func()
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
func (r *Replacer) reset() {
	if r.resetCallback != nil {
		r.resetCallback()
	}
}
func (r *Replacer) chainResetCallback(fn func()) {
	if r.resetCallback == nil {
		r.resetCallback = fn
	}
	oldCallback := r.resetCallback
	r.resetCallback = func() {
		oldCallback()
		fn()
	}
}

type Counter struct {
	c map[string]int
}

func (c *Counter) increase(prodName string) {
	c.c[prodName] = c.c[prodName] + 1
}
func (c *Counter) decrease(prodName string) {
	c.c[prodName] = c.c[prodName] - 1
}
func (c *Counter) clone() *Counter {
	ret := make(map[string]int, len(c.c))
	for k, v := range c.c {
		ret[k] = v
	}
	return &Counter{ret}
}
func (c *Counter) count(prodName string) int {
	return c.c[prodName]
}

func RandomSQLStr(productionName string, ctx *Context) []string {
	counter := &Counter{make(map[string]int)}
	defer ctx.replacer.reset()
	return randomSQLStr(productionName, ctx, counter)
}

func randomSQLStr(productionName string, ctx *Context, counter *Counter) []string {
	if debugMode {
		ctx.trace = append(ctx.trace, productionName)
		fmt.Printf("%v\n", ctx.trace)
	}

	if ctx.replacer.contains(productionName) {
		return []string{ctx.replacer.run(productionName)}
	}

	production := findProductionUnwrap(ctx.productionMap, productionName)

	seqs := filterMaxLoopAndZeroChance(production.bodyList, ctx.productionMap, counter)
	if len(seqs) == 0 {
		log.Warn("No available branch of '" + productionName + "' left")
		if debugMode {
			ctx.trace = ctx.trace[:len(ctx.trace)-1]
		}
		return nil
	}
	seqs = randomize(seqs)
	for _, s := range seqs {
		sql := generateBody(s, ctx, counter)
		if sql != nil {
			return sql
		} else {
			if debugMode {
				ctx.trace = ctx.trace[:len(ctx.trace)-len(s.seq)]
			}
		}
	}
	return nil
}

func findProductionUnwrap(prodMap map[string]*Production, name string) *Production {
	production, exist := prodMap[name]
	if !exist {
		panic(fmt.Sprintf("Production '%s' not found", name))
	}
	return production
}

func generateBody(body Body, ctx *Context, counter *Counter) []string {
	sql := []string{}
	var recordProds []string
	for _, s := range body.seq {
		if literalStr, isLiteral := literal(s); isLiteral {
			sql = appendNonEmpty(sql, literalStr)
		} else {
			fragment := randomSQLStr(s, ctx, counter.clone())
			if fragment == nil {
				log.Warn("encounter nil for " + s)
				return nil
			} else {
				sql = appendNonEmpty(sql, fragment...)
				recordProds = append(recordProds, s)
			}
		}
	}
	return sql
}

func appendNonEmpty(strs []string, adding ...string) []string {
	for _, s := range adding {
		if len(s) != 0 {
			strs = append(strs, s)
		}
	}
	return strs
}

func randomize(old BodyList) BodyList {
	bodyList := old
	totalFactor := sumRandomFactor(bodyList)
	for len(bodyList) != 0 {
		selected := pickOneBodyByRandomFactor(bodyList, totalFactor)
		bodyList[0], bodyList[selected] = bodyList[selected], bodyList[0]
		totalFactor -= bodyList[0].randomFactor
		bodyList = bodyList[1:]
	}
	return old
}

func sumRandomFactor(bodyList BodyList) int {
	total := 0
	for _, b := range bodyList {
		total += b.randomFactor
	}
	return total
}

func pickOneBodyByRandomFactor(bodyList BodyList, totalFactor int) int {
	randNum := rand.Intn(totalFactor)
	for idx, b := range bodyList {
		randNum -= b.randomFactor
		if randNum < 0 {
			return idx
		}
	}
	panic("impossible to reach")
}

func filterMaxLoopAndZeroChance(bodyList BodyList, prodMap map[string]*Production, counter *Counter) BodyList {
	var ret BodyList
	for _, body := range bodyList {
		if isZeroChance(body) || reachMaxLoop(body, prodMap, counter) {
			continue
		}
		ret = append(ret, body)
	}
	return ret
}

func reachMaxLoop(body Body, prodMap map[string]*Production, counter *Counter) bool {
	for _, s := range body.seq {
		if isLiteral(s) {
			continue
		}
		prod, ok := prodMap[s]
		if !ok {
			panic(fmt.Sprintf("production %s not found", s))
		}
		if counter.count(s) > prod.maxLoop {
			return true
		}
	}
	return false
}

func isZeroChance(body Body) bool {
	return body.randomFactor == 0
}

func literal(token string) (string, bool) {
	if isLiteral(token) {
		return strings.Trim(token, "'"), true
	}
	return "", false
}

func isLiteral(token string) bool {
	return strings.HasPrefix(token, "'") && strings.HasSuffix(token, "'")
}
