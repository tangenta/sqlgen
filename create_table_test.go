package sqlgen

import (
	"fmt"
	"strings"
	"testing"
)

func TestRandomSQL(t *testing.T) {
	ctx := buildContext()
	ss := randomSQL("create_table_stmt", nil, ctx)
	fmt.Println(strings.Join(ss, " "))
}

func TestUpdateMap(t *testing.T) {
	ret := updatedMap(map[string]int{"test": 1}, "test", increaseInt)
	if two, ok := ret["test"]; !ok || two != 2 {
		t.Fail()
	}
}
