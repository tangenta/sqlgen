module github.com/tangenta/sqlgen

go 1.12

require (
	github.com/cznic/golex v0.0.0-20181122101858-9c343928389c // indirect
	github.com/cznic/parser v0.0.0-20181122101858-d773202d5b1f // indirect
	github.com/cznic/strutil v0.0.0-20181122101858-275e90344537 // indirect
	github.com/cznic/y v0.0.0-20181122101901-b05e8c2e8d7b // indirect
	github.com/pingcap/errors v0.11.4
	github.com/pingcap/log v0.0.0-20190307075452-bd41d9273596
	github.com/pingcap/parser v0.0.0-20190805033416-34f601879210
	github.com/pingcap/tidb v3.0.0+incompatible
)

replace github.com/pingcap/parser => github.com/tangenta/parser v0.0.0-20190801031232-37156843e996
