package main

import (
	"github.com/gshirazi/crgen/pkg/crgen"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Generators for required fields
	ipnetGen, err := crgen.NewIPNetGen("100.100.100.0/24")
	if err != nil {
		log.Errorf("Could not create IPNetGen: %v", err)
		panic(err)
	}

	nhGen, err := crgen.NewSingletonGen("192.168.1.1", 1)
	if err != nil {
		log.Errorf("Could not create metricGen: %v", err)
		panic(err)
	}

	metricGen, err := crgen.NewSingletonGen("100", 1)
	if err != nil {
		log.Errorf("Could not create metricGen: %v", err)
		panic(err)
	}
	typeGen, err := crgen.NewSingletonGen("Static", 1)
	if err != nil {
		log.Errorf("Could not create typeGen: %v", err)
		panic(err)
	}
	vrfGen, err := crgen.NewSingletonGen("default", 1)
	if err != nil {
		log.Errorf("Could not create vrfGen: %v", err)
		panic(err)
	}

	myCRGen := crgen.CRGen{
		ClusterConfigPath: "/etc/rancher/k3s/k3s.yaml",
		CRDName: "routes.node.infoblox.com",
		CRDNamespace: "infoblox",
		CRDVersion: "v1",
		CRApiVersion: "node.infoblox.com/v1",
		CRKind: "Route",
		Generators: map[string]crgen.Generator{
			"metric": metricGen,
			"nextHops": nhGen,
			"prefix": ipnetGen,
			"type": typeGen,
			"vrf": vrfGen,
		},
	}
	if err := myCRGen.Generate(); err != nil {
		panic(err)
	}
}