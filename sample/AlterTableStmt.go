package sample

import (
	. "github.com/tangenta/sqlgen"
	"log"
)

var Generate = generate()

func generate() func()string {
	initState("/home/tangenta/go/src/github.com/tangenta/sqlgen/sample_bnf.txt", "AlterTableStmt")
retFn := func() string {
	res := AlterTableStmt.f()
	switch res.Tp {
	case PlainString:
		return res.Value
	case Invalid:
		log.Println("Invalid SQL")
		return ""
	case NonExist:
		log.Fatalf("Production '%s' not found", AlterTableStmt.name)
	default:
		log.Fatalf("Unsupported result type '%v'", res.Tp)
	}
	return "impossible to reach"
}


AlterTableStmt = Fn {
	name: "AlterTableStmt",
	f: func() Result {
		return random(a, Or, 
			b, 
		)
	},
}

a = Fn {
	name: "a",
	f: func() Result {
		return Str("A")
	},
}

b = Fn {
	name: "b",
	f: func() Result {
		return Str("B")
	},
}

	 return retFn
}
