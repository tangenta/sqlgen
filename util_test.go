package sqlgen

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"
)

func TestBreadthFirstSearch(t *testing.T) {
	p, err := parseYacc("mysql80_bnf_complete.txt")
	if err != nil {
		t.Error(err)
	}
	prodMap := buildProdMap(p)
	rs, err := breadthFirstSearch(prodMap, "create_table_stmt")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(len(p))
	fmt.Println(len(rs))
}

func TestParseYacc(t *testing.T) {
	p, err := parseYacc("mysql80_bnf_complete.txt")
	fmt.Printf("%v %v", p, err)
}

func TestProdSplitter(t *testing.T) {
	buf := bufio.NewReader(bytes.NewBufferString(`deallocate_or_drop: DEALLOCATE_SYM
| DROP

prepare: PREPARE_SYM ident FROM prepare_src`))
	res := splitProdStr(buf)
	for _, v := range res{
		fmt.Println(v)
	}
}

func TestIsWhitespace(t *testing.T) {
	if !isWhitespace("\t \n") {
		t.Fail()
	}
	if isWhitespace("  t  ") {
		t.Fail()
	}
	if !isWhitespace("\n  \t  ") {
		t.Fail()
	}
}