package sqlgen

import (
	"fmt"
	"strings"
	"testing"
)

func TestRandomSQL(t *testing.T) {
	prods, err := ParseYacc("create_table_config.txt")
	if err != nil {
		t.Error(err)
	}
	ctx := BuildContext(prods, buildReplacer())
	//_ = RandomSQLStr("create_table_stmt", ctx)
	//_ = RandomSQLStr("create_table_stmt", ctx)
	//ss := RandomSQLStr("create_table_stmt", ctx)
	//fmt.Println(strings.Join(ss, " "))
	for i := 0; i < 10; i++ {
		ss := RandomSQLStr("create_table_stmt", ctx)
		fmt.Println(strings.Join(ss, " "))
	}
}
