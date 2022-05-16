package main

import (
	"github.com/gshirazi/crgen/pkg/crgen"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Generators for required fields
	ipnetgen, err := crgen.NewIPNetGen("100.100.100.0/24")
	if err != nil {
		log.Errorf("Could not create IPNetGen: %v", err)
		panic(err)
	}

	myCRGen := crgen.CRGen{
		ClusterConfigPath: "/etc/rancher/k3s/k3s.yaml",
		CRDName: "routes.node.infoblox.com",
		CRDNamespace: "infoblox",
		CRDVersion: "v1",
		CRApiVersion: "node.infoblox.com/v1",
		CRKind: "Route"
		Generators: map[string]crgen.Generator{
			"metric": crgen.Singleton("100")
			"nextHops": ipnetgen,
		},
	}
	if err := myCRGen.Generate(); err != nil {
		panic(err)
	}
}