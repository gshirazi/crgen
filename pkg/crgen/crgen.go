package crgen

import (
	"context"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"strings"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	log "github.com/sirupsen/logrus"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CRGen struct {
	// ClusterConfigPath is used to get kube client
	ClusterConfigPath string
	// CRDName, CRDNamespace, CRDVersion identify the CRD
	CRDName string
	CRDNamespace string
	CRDVersion string
	// CRApiVersion and CRKind are used for generating the CRs
	CRApiVersion string
	CRKind string
	Generators map[string]Generator
}

var (
	myScheme = runtime.NewScheme()
)

func init() {
	_ = v1.AddToScheme(myScheme)
	_ = corev1.AddToScheme(myScheme)
	_ = appsv1.AddToScheme(myScheme)
}

func (c *CRGen) getKubeClient(ctx context.Context) (client.Client, error) {
	config, err := clientcmd.BuildConfigFromFlags("", c.ClusterConfigPath)
	if err != nil {
		log.Errorf("can't get kubeconfig info: %s", err)
		return nil, fmt.Errorf("can't get kubeconfig info: %s", err)
	}

	log.Debugf("get kubeconfig info: %v", config)
	// creates the kubeclient
	kubeClient, err := client.New(config, client.Options{Scheme: myScheme})
	if err != nil {
		log.Errorf("can't create kubeClient : %s", err)
		return nil, fmt.Errorf("can't create kubeClient : %s", err)
	}
	log.Debugf("create kubeClient: %v", kubeClient)
	return kubeClient, nil
}

func (c *CRGen) getCRD(ctx context.Context) (*v1.CustomResourceDefinition, error) {
	kubeClient, err := c.getKubeClient(ctx)
	if err != nil {
		return nil, err
	}
	crd := &v1.CustomResourceDefinition{}
	if err := kubeClient.Get(ctx, types.NamespacedName{Name: c.CRDName, Namespace: c.CRDNamespace}, crd); err != nil {
		return nil, err
	}
	return crd, nil
}

func (c *CRGen)getSpec() (*v1.JSONSchemaProps, error) {
	crd, err := c.getCRD(context.Background())
	if err != nil {
		return nil, err
	}
	crdVer := &v1.CustomResourceDefinitionVersion{}
	found := false
	validVersions := []string{}
	for _, ver := range crd.Spec.Versions {
		validVersions = append(validVersions, ver.Name)
		if ver.Name != c.CRDVersion {
			continue
		}
		crdVer = &ver
		found = true
	}
	if !found {
		return nil, fmt.Errorf("CRD does not have specified version %s, should be in %v", c.CRDVersion, validVersions)
	}
	spec := v1.JSONSchemaProps{}
	for name, prop := range crdVer.Schema.OpenAPIV3Schema.Properties {
		if name == "spec" {
			spec = prop
		}
	}
	log.Debugf("got crd ver %s, schema properties: %#v", c.CRDVersion, spec.Properties)
	return &spec, nil
}

type Meta struct{
	Name string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

type CRBase struct{
	ApiVersion string `yaml:"apiVersion"`
	Kind string `yaml:"kind"`
	Meta `yaml:"metadata,omitempty"`
	Spec map[string]GenValue `yaml:"spec,omitempty"`
}

func (c *CRGen)Generate() error {
	ctx := context.Background()
	_, err := c.getKubeClient(ctx)
	if err != nil {
		return err
	}

	specs, err := c.getSpec()
	if err != nil {
		return err
	}
	specNames := []string{}
	for specName := range specs.Properties {
		specNames = append(specNames, specName)
	}

	validGenerators := make(map[string]Generator)
	// validGeneratorNames := []string{}
	for g, gen := range c.Generators {
		found := false
		for _, s := range specNames {
			if s == g {
				found = true
				// validGeneratorNames = append(validGeneratorNames, g)
				validGenerators[g] = gen
				break
			}
		}
		if !found {
			log.Errorf("Generator name %s not found in CRD's specs. Skipping")
			continue
		}
	}

	crGenerator, err := NewCartesianGen(validGenerators)
	if err != nil {
		return err
	}
	crVal := crGenerator.Next()
	cnt := 0
	CR := CRBase{
		ApiVersion: c.CRApiVersion,
		Kind: c.CRKind,
	}
	for ;crVal != NIL; {
		spec := make(map[string]GenValue)
		if err := json.Unmarshal([]byte(crVal.Val), &spec); err != nil {
			log.Errorf("Unmarshal error: %v", err)
		}
		cnt = cnt + 1
		name := fmt.Sprintf("%s-crgen-%d", strings.ToLower(c.CRKind), cnt)
		CR.Meta = Meta{
			Namespace: c.CRDNamespace,
			Name: name,
		}
		CR.Spec = spec
		bytes, err := yaml.Marshal(CR)
		if err != nil {
			log.Errorf("Marshal error: %v", err)
		}
		if err = ioutil.WriteFile(fmt.Sprintf("%s.yaml", name), bytes, 0); err != nil {
			log.Errorf("Write error: %v", err)
		}
		// kubeClient.Create(ctx, obj)
		crVal = crGenerator.Next()
	}
	return nil
}
