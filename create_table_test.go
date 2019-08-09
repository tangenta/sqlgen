package sqlgen

import (
	"fmt"
	"github.com/pingcap/parser/compatibility_reporter/sql_generator"
	"strings"
	"testing"
)

func TestRandValidCreateTableStmt(t *testing.T) {
	ctx := buildContext()
	res := RandValidCreateTableStmt(100, ctx)
	fmt.Println("Valid SQL: ")
	for _, v := range res {
		fmt.Println(v)
	}

	//type kv struct {
	//	k string
	//	v int
	//}
	//var ss []kv
	//for key, value := range TempMap {
	//	ss = append(ss, kv{key, value})
	//}
	//sort.Slice(ss, func(i, j int) bool {return ss[i].v > ss[j].v})
	//for _, kv := range ss {
	//	fmt.Printf("key: %v, value: %v\n", kv.k, kv.v)
	//}
}

func TestEnumExpression(t *testing.T) {
	//_ = sql_flat_generator.NewSQLEnumIterator(parseProductions(), "expr")
	//for i := 0; i < 1000 && iter.HasNext(); i++ {
	//	fmt.Println(iter.Next())
	//}
}

func TestRandExpr(t *testing.T) {
	iter := sql_generator.GenerateSQLRandomly(parseProductions(), "expr")
	for i := 0; i < 100; i++ {
		fmt.Println(iter.Next())
	}
}

func TestRandomSQL(t *testing.T) {
	prods := parseProductions()
	prodMap := initProductionMap(prods)
	cfg := buildContext().randConfig
	ss := randomSQL("table_element_list", nil, prodMap, cfg)
	fmt.Println(strings.Join(ss, " "))
}