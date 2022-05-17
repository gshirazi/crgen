package crgen

import (
	"encoding/json"
	"sort"

	"github.com/korylprince/ipnetgen"
)

const (
	NIL = ""
)

type Generator interface {
	Reset()
	Next() string
}

type SingletonGen struct {
	Val string
	Count uint32
	cnt uint32
}

func NewSingletonGen(val string, cnt uint32) (*SingletonGen, error) {
	return &SingletonGen{Val: val, Count: cnt}, nil
}

func (g *SingletonGen) Reset() {
	g.cnt = 0
}

func (g *SingletonGen) Next() string {
	if g.cnt < g.Count {
		g.cnt = g.cnt + 1
		return g.Val
	}
	return NIL
}

type IPNetGen struct {
	ipgen *ipnetgen.IPNetGenerator
}

func NewIPNetGen(subnet string) (*IPNetGen, error) {
	ipgen , err := ipnetgen.New(subnet)
	if err != nil {
		return nil, err
	}
	return &IPNetGen{
		ipgen: ipgen,
	}, nil
}

func (g *IPNetGen) Reset() {
	g.ipgen = ipnetgen.NewFromIPNet(g.ipgen.IPNet)
}

func (g *IPNetGen) Next() string {
	nextIp := g.ipgen.Next()
	if nextIp == nil {
		return NIL
	}
	return nextIp.String()
}

type CartesianGen struct {
	Fields map[string]Generator
	current map[string]string
	fieldNamesSorted []string
}

func NewCartesianGen(fields map[string]Generator) (*CartesianGen, error) {
	return &CartesianGen{Fields: fields}, nil
}

func (g *CartesianGen) init() {
	g.current = make(map[string]string)
	fieldNames := make([]string, 0, len(g.Fields))
	for f := range g.Fields {
		fieldNames = append(fieldNames, f)
	}
	sort.Strings(fieldNames)
	g.fieldNamesSorted = fieldNames
}

func (g *CartesianGen) Reset() {
	for _, gen := range g.Fields {
		gen.Reset()
	}
	g.current = nil
}

func (g *CartesianGen) Next() string {
	if g.current == nil {
		g.init()
		for _, f := range g.fieldNamesSorted {
			g.current[f] = g.Fields[f].Next()
		}
		bytes, _ := json.Marshal(g.current)
		return string(bytes)
	}
	for _, f := range g.fieldNamesSorted {
		nextVal := g.Fields[f].Next()
		if nextVal != NIL {
			g.current[f] = nextVal
			bytes, _ := json.Marshal(g.current)
			return string(bytes)
		}
		g.Fields[f].Reset()
		g.current[f] = g.Fields[f].Next()
	}
	return NIL
}
