package sqlgen

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"
)

//func buildConfigFile(prodName string, prodMap map[string]*Production) string {
//	var res string
//
//}

func buildProdMap(prods []*Production) map[string]*Production {
	ret := make(map[string]*Production)
	for _, v := range prods {
		ret[v.head] = v
	}
	return ret
}

func breadthFirstSearch(prodMap map[string]*Production, prodName string) (map[string]struct{}, error) {
	resultSet := map[string]struct{}{}
	pendingSet := []string{prodName}

	for len(pendingSet) != 0 {
		name := pendingSet[0]
		pendingSet = pendingSet[1:]
		prod, ok := prodMap[name]
		if !ok {
			return nil, fmt.Errorf("%v not found", name)
		}
		if _, contains := resultSet[name]; !contains {
			resultSet[name] = struct{}{}
			for _, body := range prod.bodyList {
				for _, s := range body.seq{
					if _, isLit := literal(s); !isLit {
						pendingSet = append(pendingSet, s)
					}
				}
			}
		}
	}
	return resultSet, nil
}

func union(map1 map[string]struct{}, map2 map[string]struct{}) map[string]struct{} {
	for key, value := range map2 {
		map1[key] = value
	}
	return map1
}

func binSearch(sortedProds []*Production, prodName string) *Production {
	idx := sort.Search(len(sortedProds), func(i int) bool { return sortedProds[i].head >= prodName})
	if idx < len(sortedProds) && sortedProds[idx].head == prodName {
		return sortedProds[idx]
	}
	return nil
}

func parseYacc(yaccFilePath string) ([]*Production, error) {
	file, err := os.Open(yaccFilePath)
	if err != nil {
		return nil, err
	}

	prodStrs := splitProdStr(bufio.NewReader(file))
	return parseProdStr(prodStrs)
}

func parseProdStr(prodStrs []string) ([]*Production, error) {
	bnfParser := New()
	var ret []*Production
	for _, p := range prodStrs {
		r, _, err := bnfParser.Parse(p)
		if err != nil {
			return nil, err
		}
		ret = append(ret, r)
	}
	return ret, nil
}

func splitProdStr(prodReader *bufio.Reader) []string {
	var ret []string
	var sb strings.Builder
	time2Exit := false
	for !time2Exit {
		for {
			str, err := prodReader.ReadString('\n')
			if err != nil {
				time2Exit = true
				if !isWhitespace(str) {
					sb.WriteString(str)
				}
				break
			}
			if isWhitespace(str) && sb.Len() != 0 {
				ret = append(ret, sb.String())
				sb.Reset()
			} else {
				sb.WriteString(str)
			}
		}
	}
	if sb.Len() != 0 {
		ret = append(ret, sb.String())
	}
	return ret
}

func isWhitespace(str string) bool {
	for _, c := range str {
		if !unicode.IsSpace(c) {
			return false
		}
	}
	return true
}
