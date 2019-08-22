package sqlgen

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
)

func buildFile(yaccFilePath, prodName, packageName, outputFilePath string) {
	yaccFilePath = explicitPath(yaccFilePath)
	outputFilePath = explicitPath(outputFilePath)
	prods, err := ParseYacc(yaccFilePath)
	if err != nil {
		log.Fatal(err)
	}
	prodMap := BuildProdMap(prods)

	Must(os.Chdir(outputFilePath))
	Must(os.Mkdir(packageName, 0755))
	Must(os.Chdir(packageName))
	oFile, err := os.OpenFile(prodName + ".go", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	defer func() { _ = oFile.Close() }()
	if err != nil {
		log.Fatal(err)
	}

	MustWrite(oFile, packageDirective())
	MustWrite(oFile, importDirective)

	MustWrite(oFile, generateDirective)
	MustWrite(oFile, fmt.Sprintf("\nfunc generate() func()string {"))
	MustWrite(oFile, pubInterface(yaccFilePath, prodName))

	visitor := func(p *Production) {
		MustWrite(oFile, convertProdToCode(p))
	}
	allProds, err := breadthFirstSearch(prodName, prodMap, visitor)
	if err != nil {
		log.Fatal(err)
	}
	MustWrite(oFile, "\n\t return retFn\n}\n")

	utilFile, err := os.OpenFile("util.go", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	defer func() { _ = utilFile.Close() }()
	if err != nil {
		log.Fatal(err)
	}
	MustWrite(utilFile, packageDirective())
	MustWrite(utilFile, utilSnippet)

	declareFile, err := os.OpenFile("declarations.go", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	defer func() { _ = declareFile.Close() }()
	if err != nil {
		log.Fatal(err)
	}
	MustWrite(declareFile, packageDirective())
	for p := range allProds {
		p = convertHead(p)
		MustWrite(declareFile, convertNameToDeclaration(p))
	}

	testFile, err := os.OpenFile(packageName + "_test.go", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	defer func() { _ = testFile.Close() }()
	if err != nil {
		log.Fatal(err)
	}
	MustWrite(testFile, packageDirective())
	MustWrite(testFile, testSnippet)
}

func packageDirective() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal("Cannot get working directory")
	}
	dirs := strings.Split(dir, "/")
	packageName := dirs[len(dirs)-1]
	return fmt.Sprintf("package %s\n", packageName)
}

const importDirective = `
import (
	. "github.com/tangenta/sqlgen"
	"log"
)
`

const generateDirective = `
var Generate = generate()
`

const utilSnippet = `
import (
	. "github.com/tangenta/sqlgen"
	"log"
	"math/rand"
	"strings"
	"time"
)

var state = State{
	Choices:           nil,
	Counter:           map[string]int{},
	TotalCounter:      map[string]int{},
	CurrentProduction: nil,

	ProductionMap:       nil,
	BeginProductionName: "",
	IsInitialize:        false,
}

// Fn is able to manipulate global state, simulating calling stack.
type Fn struct {
	name        string
	f           func() Result
	isBranchTag bool // Mark for splitter '|'.
}

func (fn *Fn) callWithLoc(branchNum, SeqNum int) Result {
	state.EnsureInitialized()
	if fn.isBranchTag {
		log.Fatal("Cannot call on Branch tag")
	}

	choice := Choice{Branch: branchNum, SeqNum: SeqNum}
	state.Choices = append(state.Choices, choice)

	fnName := fn.name
	// Before calling function.
	state.Counter[fnName] += 1
	state.TotalCounter[fnName] += 1
	state.CurrentProduction = findProductionAndUnwrap(fnName, state.ProductionMap)

	ret := fn.f()
	// After calling function.
	parent := state.Parent()
	state.Choices = state.Choices[:len(state.Choices)-1]
	state.Counter[fnName] -= 1
	state.CurrentProduction = parent
	return ret
}

func (fn *Fn) discard() {
	state.EnsureInitialized()

	fnName := fn.name
	state.Counter[fnName] -= 1
	state.TotalCounter[fnName] -= 1
}

// ----- utilities ------

func random(symbols ...Fn) Result {
	branches := splitBranches(symbols)
	return randomBranch(branches)
}

func randomBranch(branches [][]Fn) Result {
	branchNum := len(branches)
	if branchNum <= 0 {
		return Result{Tp: Invalid}
	}
	chosenBranchNum := rand.Intn(branchNum)
	chosenBranch := branches[chosenBranchNum]

	var doneF []Fn
	var resStr strings.Builder
	for i, f := range chosenBranch {
		res := f.callWithLoc(chosenBranchNum, i)
		switch res.Tp {
		case PlainString:
			doneF = append(doneF, f)
			if i != 0 {
				resStr.WriteString(" ")
			}
			resStr.WriteString(res.Value)
		case NonExist:
			log.Fatalf("Production '%s' not found", f.name)
		case Invalid:
			for _, df := range doneF {
				df.discard()
			}
			branches[chosenBranchNum], branches[0] = branches[0], branches[chosenBranchNum]
			return randomBranch(branches[1:])
		default:
			log.Fatalf("Unsupported result type '%v'", res.Tp)
		}
	}
	return Str(resStr.String())
}

func splitBranches(fns []Fn) [][]Fn {
	var ret [][]Fn
	var Branch []Fn
	for _, f := range append(fns, Or) {
		if f.isBranchTag {
			if len(Branch) == 0 {
				log.Fatal("Empty Branch is impossible to split")
			}
			ret = append(ret, Branch)
			Branch = nil
		} else {
			Branch = append(Branch, f)
		}
	}
	return ret
}

func findProductionAndUnwrap(name string, prodMap map[string]*Production) *Production {
	ret, ok := prodMap[name]
	if !ok {
		log.Fatalf("Production '%s' not found", name)
	}
	return ret
}

func initState(bnfFileName string, beginProdName string) {
	prods, err := ParseYacc(bnfFileName)
	if err != nil {
		log.Fatal(err)
	}
	prodMap := BuildProdMap(prods)
	beginProd, ok := prodMap[beginProdName]
	if !ok {
		log.Fatalf("Begin production name '%s' not found", beginProdName)
	}

	state.ProductionMap = prodMap
	state.CurrentProduction = beginProd
	state.BeginProductionName = beginProdName
	rand.Seed(time.Now().UnixNano())
	state.IsInitialize = true
}

func constFn(str string) Fn {
	return Fn{name: str, f: func() Result {
		return Result{Tp:PlainString, Value: str}
	}}
}

func Str(str string) Result {
	return Result{Tp: PlainString, Value: str}
}

var Or = Fn{isBranchTag: true}

`

const templateDriver = `
	initState("%s", "%s")
retFn := func() string {
	res := %s.f()
	switch res.Tp {
	case PlainString:
		return res.Value
	case Invalid:
		log.Println("Invalid SQL")
		return ""
	case NonExist:
		log.Fatalf("Production '%%s' not found", %s.name)
	default:
		log.Fatalf("Unsupported result type '%%v'", res.Tp)
	}
	return "impossible to reach"
}

`

func pubInterface(yaccFilePath, prodName string) string {
	return fmt.Sprintf(templateDriver, yaccFilePath, prodName, prodName, prodName)
}

const templateR = `
%s = Fn {
	name: "%s",
	f: func() Result {
		return random(%s
		)
	},
}
`

const templateS = `
%s = Fn {
	name: "%s",
	f: func() Result {
		return Str(%s)
	},
}
`

func convertProdToCode(p *Production) string {
	prodHead := convertHead(p.head)
	if len(p.bodyList) == 1 {
		allLiteral := true
		seqs := p.bodyList[0].seq
		for _, s := range seqs {
			if !isLiteral(s) {
				allLiteral = false
				break
			}
		}

		trimmedSeqs := trimmedStrs(seqs)
		if allLiteral {
			return fmt.Sprintf(templateS, prodHead, prodHead, strings.Join(trimmedSeqs, " "))
		}
	}

	var bodyStr strings.Builder
	for i, body := range p.bodyList {
		for _, s := range body.seq {
			if isLit, ok := literal(s); ok {
				s = fmt.Sprintf("constFn(\"%s\")", isLit)
			} else {
				s = convertHead(s)
			}
			bodyStr.WriteString(s)
			bodyStr.WriteString(", ")
		}
		if i != len(p.bodyList)-1 {
			bodyStr.WriteString("Or, \n\t\t\t")
		}
	}

	return fmt.Sprintf(templateR, prodHead, p.head, bodyStr.String())
}

const templateDecl = "var %s Fn\n"

func convertNameToDeclaration(name string) string {
	return fmt.Sprintf(templateDecl, name)
}

func trimmedStrs(origin []string) []string {
	ret := make([]string, len(origin))
	for i, s := range origin {
		if lit, ok := literal(s); ok {
			ret[i] = fmt.Sprintf("\"%s\"", lit)
		}
	}
	return ret
}

// convertHead to avoid keyword clash.
func convertHead(str string) string {
	if strings.HasPrefix(str, "$@") {
		return "num" + strings.TrimPrefix(str, "$@")
	}

	switch str {
	case "type": return "utype"
	case "%empty": return "empty"
	default: return str
	}
}

func MustWrite(oFile *os.File, str string) {
	_, err := oFile.WriteString(str)
	if err != nil {
		log.Fatal(err)
	}
}

func Must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func explicitPath(p string) string {
	if !strings.ContainsRune(p, os.PathSeparator) {
		dir, err := os.Getwd()
		if err != nil {
			log.Fatal("Unable to get pwd")
		}
		return path.Join(dir, p)
	}
	return p
}

const testSnippet = `
import (
	"fmt"
	"testing"
)

func TestA(t *testing.T) {
	for i := 0; i < 10; i++ {
		fmt.Println(Generate())
	}
}
`