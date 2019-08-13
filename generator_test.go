package sqlgen

import (
	"fmt"
	"strings"
	"testing"
)

func TestRandomSQL(t *testing.T) {
	prods, err := parseYacc("create_table_config.txt")
	if err != nil {
		t.Error(err)
	}
	ctx := BuildContext(prods, buildReplacer())
	for i := 0; i < 10; i++ {
		ss := RandomSQLStr("create_table_stmt", ctx)
		fmt.Println(strings.Join(ss, " "))
	}
}
