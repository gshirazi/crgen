# crgen
Generate multiple CR instances for a CRD

# Usage

Assumming you have a CRD `my.crd.full.name/MyCRD` in `mynamespace` with 4 required fields `spec1,2,3,4`;
The following code snippet creates `255 * 1 * 5 * 1 = 1275` CR yaml files that can be installed on the cluster (using `kubectl apply -f`) for stress tests:

```
import "github.com/gshirazi/crgen/pkg/crgen"

spec1Gen, err := crgen.NewIPNetGen("100.100.100.0/24")
if err != nil {
    log.Errorf("Could not create spec1Gen: %v", err)
    panic(err)
}

spec2Gen, err := crgen.NewSingletonGen("elem1,elem2", 1, "array")
if err != nil {
    log.Errorf("Could not create spec2Gen: %v", err)
    panic(err)
}

spec3Gen, err := crgen.NewSingletonGen("100", 5, "int")
if err != nil {
    log.Errorf("Could not create spec3Gen: %v", err)
    panic(err)
}

spec4Gen, err := crgen.NewSingletonGen("mystring", 1, "string")
if err != nil {
    log.Errorf("Could not create spec4Gen: %v", err)
    panic(err)
}

func main() {
	myCRGen := crgen.CRGen{
		ClusterConfigPath: "/etc/rancher/k3s/k3s.yaml",
		CRDName: "my.crd.full.name",
		CRDNamespace: "mynamespace",
		CRDVersion: "v1",
		CRApiVersion: "gshirazi.github.com/v1",
		CRKind: "MyCRD",
		Generators: map[string]crgen.Generator{
			"spec1": spec1Gen,
			"spec2": spec2Gen,
			"spec3": spec3Gen,
			"spec4": spec4Gen,
		},
	}
	if err := myCRGen.Generate(); err != nil {
		panic(err)
	}
}
```
As it can be seen, `crgen` supports iterating over IPs, and is also capable of generating `int`, `array` and `string` types. A custom type can be added by implementing `GenVal` and its `MarshalYAML()`.
