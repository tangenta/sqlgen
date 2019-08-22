package sqlgen

import "testing"

func TestNewGenerator(t *testing.T) {
	buildFile("sample_bnf.txt", "AlterTableStmt", "sample",".")
}

func TestFullGenerator(t *testing.T) {
	buildFile("mysql80_bnf_complete.txt", "alter_table_stmt", "alter", ".")
}
