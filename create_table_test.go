package sqlgen

import (
	"fmt"
	"strings"
	"testing"
)

func TestRandomSQL(t *testing.T) {
	prods := parseProductions()
	prodMap := initProductionMap(prods)
	cfg := buildContext().randConfig
	ss := randomSQL("table_element_list", nil, prodMap, cfg)
	fmt.Println(strings.Join(ss, " "))
}