.PHONY: all clean

ARCH:="`uname -s`"
MAC:="Darwin"
LINUX:="Linux"

all: bnf
	@echo "make all"

bnf: bin/goyacc
	bin/goyacc -o /dev/null bnf.y
	bin/goyacc -o bnf_parser.go bnf.y 2>&1 | egrep "(shift|reduce)/reduce" | awk '{print} END {if (NR > 0) {print "Find conflict in parser.y. Please check y.output for more information."; exit 1;}}'
	rm -f y.output

	@if [ $(ARCH) = $(LINUX) ]; \
	then \
		sed -i -e 's|//line.*||' -e 's/yyEofCode/yyEOFCode/' bnf_parser.go; \
	elif [ $(ARCH) = $(MAC) ]; \
	then \
		/usr/bin/sed -i "" 's|//line.*||' bnf_parser.go; \
		/usr/bin/sed -i "" 's/yyEofCode/yyEOFCode/' bnf_parser.go; \
	fi

	@awk 'BEGIN{print "// Code generated by goyacc DO NOT EDIT."} {print $0}' bnf_parser.go > tmp_parse_bnf.go && mv tmp_parse_bnf.go bnf_parser.go;

bin/goyacc: goyacc/main.go
	GO111MODULE=on go build -o bin/goyacc goyacc/main.go

fmt:
	@echo "gofmt (simplify)"
	@ gofmt -s -l -w . 2>&1 | awk '{print} END{if(NR>0) {exit 1}}'

clean:
	go clean -i ./...
	echo "make clean"
