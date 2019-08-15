package sqlgen

import (
	"fmt"
	"github.com/pingcap/log"
	"math/rand"
	"strings"
)

type Context struct {
	productionMap map[string]*Production
	replacer      *Replacer
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

func RandomSQLStr(productionName string, ctx *Context) []string {
	counter := make(map[string]int)
	defer func() {
		for key, count := range counter {
			ctx.productionMap[key].maxLoop += count
		}
	}()
	return randomSQLStr(productionName, ctx, counter)
}

func randomSQLStr(productionName string, ctx *Context, counter map[string]int) []string {
	if ctx.replacer.contains(productionName) {
		return []string{ctx.replacer.run(productionName)}
	}

	production := findProductionUnwrap(ctx.productionMap, productionName)
	production.maxLoop -= 1
	counter[productionName] += 1

	seqs := filterMaxLoopAndZeroChance(production.bodyList, ctx.productionMap)
	if len(seqs) == 0 {
		log.Debug("exiting from " + productionName)
		return nil
	}
	seqs = randomize(seqs)
	for _, s := range seqs {
		sql := generateBody(s, ctx, counter)
		if sql != nil {
			return sql
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

func generateBody(body Body, ctx *Context, counter map[string]int) []string {
	sql := []string{}
	for _, s := range body.seq {
		if literalStr, isLiteral := literal(s); isLiteral {
			sql = appendNonEmpty(sql, literalStr)
		} else {
			fragment := randomSQLStr(s, ctx, counter)
			if fragment == nil {
				log.Debug("encounter nil for " + s)
				return nil
			} else {
				sql = appendNonEmpty(sql, fragment...)
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

func filterMaxLoopAndZeroChance(bodyList BodyList, prodMap map[string]*Production) BodyList {
	var ret BodyList
	for _, body := range bodyList {
		if isZeroChance(body) || reachMaxLoop(body, prodMap) {
			continue
		}
		ret = append(ret, body)
	}
	return ret
}

func reachMaxLoop(body Body, prodMap map[string]*Production) bool {
	for _, s := range body.seq {
		if isLiteral(s) {
			continue
		}
		prod, ok := prodMap[s]
		if !ok {
			panic(fmt.Sprintf("production %s not found", s))
		}
		if prod.maxLoop <= 0 {
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
