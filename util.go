package sqlgen

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
)

func buildConfigFile(prodName string, prodMap map[string]*Production, filePath string) error {
	oFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	_, writeErr := oFile.WriteString(convertOrigin(prodName, prodMap))
	if writeErr != nil {
		return writeErr
	}
	return nil
}

func convertOrigin(prodName string, prodMap map[string]*Production) string {
	var sb strings.Builder
	visitor := func(p *Production) {
		sb.WriteString(p.String())
		sb.WriteString("\n")
	}
	_, _ = breadthFirstSearch(prodName, prodMap, visitor)

	return sb.String()
}

func BuildProdMap(prods []*Production) map[string]*Production {
	ret := make(map[string]*Production)
	for _, v := range prods {
		ret[v.head] = v
	}
	checkProductionMap(ret)
	return ret
}

func checkProductionMap(productionMap map[string]*Production) {
	for _, production := range productionMap {
		for _, seqs := range production.bodyList {
			for _, seq := range seqs.seq {
				if isLiteral(seq) {
					continue
				}
				if _, exist := productionMap[seq]; !exist {
					panic(fmt.Sprintf("Production '%s' not found", seq))
				}
			}
		}
	}
}

func breadthFirstSearch(prodName string, prodMap map[string]*Production, visitors ...func(*Production)) (map[string]struct{}, error) {
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
				for _, s := range body.seq {
					if !isLiteral(s) {
						pendingSet = append(pendingSet, s)
					}
				}
			}
			if len(visitors) != 0 {
				for _, v := range visitors {
					v(prod)
				}
			}
		}
	}
	return resultSet, nil
}

func ParseYacc(yaccFilePath string) ([]*Production, error) {
	file, err := os.Open(yaccFilePath)
	if err != nil {
		return nil, err
	}

	prodStrs := splitProdStr(bufio.NewReader(file))
	return parseProdStr(prodStrs)
}

func parseProdStr(prodStrs []string) ([]*Production, error) {
	bnfParser := NewParser()
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
