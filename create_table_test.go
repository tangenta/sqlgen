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
	ctx := buildContext(prods)
	ss := randomSQL("create_table_stmt", ctx)
	fmt.Println(strings.Join(ss, " "))
}
