package crgen

import (
	"encoding/json"
	"github.com/korylprince/ipnetgen"
)

const (
	NIL = ""
)

type Generator interface {
	Next() string
}

type SingletonGen struct {
	Val string
	Count uint32
}

func NewSingletonGen(val string, cnt uint32) (*SingletonGen, error) {
	return &SingletonGen{Val: val, Count: cnt}, nil
}

func (g *SingletonGen) Next() string {
	if g.Count {
		g.Count = g.Count - 1
		return Val
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
	indices map[string]uint32
	initFields map[string]Generator
	lastFieldName string
}

func NewCartesianGen(fileds map[string]Generator) (*CartesianGen, error) {
	return &CartesianGen{Fields: fields}, nil
}

func (g *CartesianGen) initIndices() {
	for f := range Fields {
		indices[f] := 0
		initFields[f] = Fields[f]
		lastFieldName = f
	}
}

func (g *CartesianGen) Next() string {
	v = map[string]string
	if indices == nil {
		// first time, initialize indices
		g.initIndices()
		for f, gen := range Fields {
			g.current[f] = gen.Next()
		}
		bytes, _ := json.Marshal(g.current)
		return string(bytes)
	}
	for f, gen := range Fields {
		nextVal := gen.Next()
		if nextVal != NIL {
			g.current[f] = nextVal
			bytes, _ := json.Marshal(g.current)
			return string(bytes)
		}
		/*if f == lastFieldName {
			return NIL
		}*/
		Fields[f] = initFields[f]
		g.current[f] = Fields[f].Next()
	}
	return NIL
}