package crgen

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"github.com/korylprince/ipnetgen"
)

var (
	NIL = GenValue{Type: "string", Val: "NIL"}
)

type Generator interface {
	Reset()
	Next() GenValue
}

type GenValue struct {
	Type string
	Val string
}

// TODO: custom types should be implemented from outside library
type Prefix struct {
	Family string `yaml:"family"`
	IPv4 string `yaml:"ipv4"`
}

func (v GenValue) MarshalYAML() (interface{}, error) {
	switch v.Type {
	case "string","ip":
		return v.Val, nil
	case "int":
		intVal, err := strconv.Atoi(v.Val)
		if err != nil {
			return 0, err
		}
		return intVal, nil
	case "array":
		return strings.Split(v.Val, ","), nil
	// TODO: custom types should be implemented from outside library
	case "prefix":
		return Prefix{
			Family: "IPv4",
			IPv4: v.Val,
		}, nil
	default:
		return v.Val, nil
	}
	return v.Val, nil
}

type SingletonGen struct {
	Val string
	Count uint32
	Type string
	cnt uint32
}

func NewSingletonGen(val string, cnt uint32, typ string) (*SingletonGen, error) {
	return &SingletonGen{Val: val, Count: cnt, Type: typ}, nil
}

func (g *SingletonGen) Reset() {
	g.cnt = 0
}

func (g *SingletonGen) Next() GenValue {
	if g.cnt < g.Count {
		g.cnt = g.cnt + 1
		return GenValue{Type: g.Type, Val: g.Val}
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

func (g *IPNetGen) Next() GenValue {
	nextIp := g.ipgen.Next()
	if nextIp == nil {
		return NIL
	}
	return GenValue{Val: nextIp.String(), Type: "prefix"} // Type: "ip"
}

type CartesianGen struct {
	Fields map[string]Generator
	current map[string]GenValue
	fieldNamesSorted []string
}

func NewCartesianGen(fields map[string]Generator) (*CartesianGen, error) {
	return &CartesianGen{Fields: fields}, nil
}

func (g *CartesianGen) init() {
	g.current = make(map[string]GenValue)
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

func (g *CartesianGen) Next() GenValue {
	if g.current == nil {
		g.init()
		for _, f := range g.fieldNamesSorted {
			g.current[f] = g.Fields[f].Next()
		}
		bytes, _ := json.Marshal(g.current)
		return GenValue{Val: string(bytes), Type: "json"}
	}
	for _, f := range g.fieldNamesSorted {
		nextVal := g.Fields[f].Next()
		if nextVal != NIL {
			g.current[f] = nextVal
			bytes, _ := json.Marshal(g.current)
			return GenValue{Val: string(bytes), Type: "json"}
		}
		g.Fields[f].Reset()
		g.current[f] = g.Fields[f].Next()
	}
	return NIL
}
