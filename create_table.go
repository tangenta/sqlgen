package sqlgen

import (
	"bufio"
	"fmt"
	"github.com/pingcap/parser"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/pingcap/parser/compatibility_reporter/yacc_parser"
	_ "github.com/pingcap/tidb/types/parser_driver"
)

type Context struct {
	productions []yacc_parser.Production
	tidbParser *parser.Parser
	randConfig *RandConfig
}
func (c *Context) Clone() Context {
	rc := c.randConfig.Clone()
	return Context{
		productions: c.productions,
		tidbParser:  c.tidbParser,
		randConfig:  &rc,
	}
}

type RandConfig struct {
	maxLoopback int
	loopBackWhiteList map[string]struct{}
	// The production name or literal that contains strBlackList is skipped.
	strBlackList map[string]struct{}
	replacer *Replacer
}
func (r *RandConfig) Clone() RandConfig {
	re := r.replacer.Clone()
	return RandConfig{
		maxLoopback:       r.maxLoopback,
		loopBackWhiteList: copySet(r.loopBackWhiteList),
		strBlackList:      copySet(r.strBlackList),
		replacer:          &re,
	}
}
func copySet(oldMap map[string]struct{}) map[string]struct{} {
	newMap := make(map[string]struct{}, len(oldMap))
	for k, v := range oldMap {
		newMap[k] = v
	}
	return newMap
}

type Replacer struct {
	s map[string]func()string
}
func (r *Replacer) add(str string, strSupplier func()string) {
	if r.s == nil {
		r.s = make(map[string]func()string)
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
func (r *Replacer) Clone() Replacer {
	newMap := make(map[string]func()string, len(r.s))
	for k, v := range r.s {
		newMap[k] = v
	}
	return Replacer{s: newMap}
}

func buildContext() (ctx *Context) {
	ctx = &Context{
		productions: parseProductions(),
		tidbParser:  parser.New(),
		randConfig: &RandConfig{
			maxLoopback:       3,
			loopBackWhiteList: buildStringSet(),
			strBlackList: buildStringSet(),
			replacer: &Replacer{},
		},
	}
	return
}

func constStrFn(str string) func() string {
	return func() string { return str}
}

func buildStringSet(args ...string) map[string]struct{} {
	result := make(map[string]struct{}, len(args))
	for _, v := range args{
		result[v] = struct{}{}
	}
	return result
}

func parseProductions() []yacc_parser.Production {
	bnfs := []string{"mysql80_bnf_complete.txt", "mysql80_custom.txt", "mysql80_lexical.txt"}
	var allProductions []yacc_parser.Production
	for _, bnf := range bnfs {
		bnfFile, err := os.Open(bnf)
		if err != nil {
			panic(fmt.Sprintf("File '%s' open failure", bnf))
		}
		productions := yacc_parser.Parse(yacc_parser.Tokenize(bufio.NewReader(bnfFile)))
		allProductions = append(allProductions, productions...)
	}
	return allProductions
}

func initProductionMap(productions []yacc_parser.Production) map[string]yacc_parser.Production {
	productionMap := make(map[string]yacc_parser.Production)
	for _, production := range productions {
		if pm, exist := productionMap[production.Head]; exist {
			pm.Alter = append(pm.Alter, production.Alter...)
			productionMap[production.Head] = pm
		}
		productionMap[production.Head] = production
	}
	checkProductionMap(productionMap)
	return productionMap
}
func checkProductionMap(productionMap map[string]yacc_parser.Production) {
	for _, production := range productionMap {
		for _, seqs := range production.Alter {
			for _, seq := range seqs.Items {
				if _, isLiteral := literal(seq); isLiteral {
					continue
				}
				if _, exist := productionMap[seq]; !exist {
					panic(fmt.Sprintf("Production '%s' not found", seq))
				}
			}
		}
	}
}

var nothing []string

func randomSQL(productionName string, counter map[string]int, productionMap map[string]yacc_parser.Production, cfg *RandConfig) []string {
	production, exist := productionMap[productionName]
	if !exist {
		panic(fmt.Sprintf("Production '%s' not found", productionName))
	}

	seqs := filterMaxLoopback(production.Alter, counter, cfg.maxLoopback)
	if len(seqs) == 0 {
		return nothing
	}
	seqs = randomize(seqs)
	for _, s := range seqs {
		var sql []string
		containsException := false
		for _, item := range s.Items {
			if literalStr, isLiteral := literal(item); isLiteral {
				if literalStr != "" {
					sql = append(sql, literalStr)
				}
			} else {
				fragment := randomSQL(item, updatedMap(counter, productionName, increaseInt), productionMap, cfg)
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

func increaseInt(a int) int {
	return a + 1
}

func updatedMap(old map[string]int, key string, updFn func(int)int) map[string]int {
	ret := make(map[string]int, len(old) + 1)

	updated := false
	for oKey, oValue := range old {
		if oKey != key {
			ret[oKey] = oValue
		} else {
			ret[key] = updFn(ret[key])
			updated = true
		}
	}
	if !updated {
		ret[key] = 1
	}
	return ret
}

func randomize(old []yacc_parser.Seq) []yacc_parser.Seq {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(old), func (i, j int) {old[i], old[j] = old[j], old[i]})
	return old
}

func filterMaxLoopback(seq []yacc_parser.Seq, counter map[string]int, maxLoopback int) []yacc_parser.Seq {
	var ret []yacc_parser.Seq
	for _, v := range seq {
		exceedMax := false
		for _, i := range v.Items {
			count, ok := counter[i]
			if !ok {
				count = 0
			}
			if count > maxLoopback {
				exceedMax = true
				break
			}
		}
		if !exceedMax {
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