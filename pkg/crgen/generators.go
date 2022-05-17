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
	if g.Count != 0 {
		g.Count = g.Count - 1
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

func NewCartesianGen(fields map[string]Generator) (*CartesianGen, error) {
	return &CartesianGen{Fields: fields}, nil
}

func (g *CartesianGen) initIndices() {
	g.indices = make(map[string]uint32)
	g.initFields = make(map[string]Generator)
	g.current = make(map[string]string)
	for f := range g.Fields {
		g.indices[f] = 0
		g.initFields[f] = g.Fields[f]
		g.lastFieldName = f
	}
}

func (g *CartesianGen) Next() string {
	// v := make(map[string]string)
	if g.indices == nil {
		// first time, initialize indices
		g.initIndices()
		for f, gen := range g.Fields {
			g.current[f] = gen.Next()
		}
		bytes, _ := json.Marshal(g.current)
		return string(bytes)
	}
	for f, gen := range g.Fields {
		nextVal := gen.Next()
		if nextVal != NIL {
			g.current[f] = nextVal
			bytes, _ := json.Marshal(g.current)
			return string(bytes)
		}
		/*if f == lastFieldName {
			return NIL
		}*/
		g.Fields[f] = g.initFields[f]
		g.current[f] = g.Fields[f].Next()
	}
	return NIL
}