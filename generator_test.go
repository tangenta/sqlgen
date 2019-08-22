package sqlgen

import "testing"

func TestNewGenerator(t *testing.T) {
	buildFile("sample_bnf.txt", "start", "sample",".")
}
