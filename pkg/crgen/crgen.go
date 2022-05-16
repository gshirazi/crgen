package crgen

import (
	"context"
	"fmt"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	log "github.com/sirupsen/logrus"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	// kubernetes/apiextensions-apiserver
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
	scheme = runtime.NewScheme()
)

func init() {
	_ = v1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
}

func (c *CRGen) getKubeClient(ctx context.Context) (client.Client, error) {
	config, err := clientcmd.BuildConfigFromFlags("", c.ClusterConfigPath)
	if err != nil {
		log.Errorf("can't get kubeconfig info: %s", err)
		return nil, fmt.Errorf("can't get kubeconfig info: %s", err)
	}

	log.Debugf("get kubeconfig info: %v", config)
	// creates the kubeclient
	kubeClient, err := client.New(config, client.Options{Scheme: scheme})
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

func (c *CRGen)getSpecNames() ([]string, error) {
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

	specNames := []string{}
	for specName := range spec.Properties {
		specNames = append(specNames, specName)
	}
	return specNames, nil
}

func (c *CRGen)Generate() error {
	ctx := context.Background()
	kubeClient, err := c.getKubeClient(ctx)
	if err != nil {
		return err
	}

	specs, err := c.getSpecNames()
	if err != nil {
		return err
	}

	validGenerators := make(map[string]Generator)
	// validGeneratorNames := []string{}
	for g, gen := range c.Generators {
		found := false
		for _, s := range specs {
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
	CR := map[string]string{
		"apiVersion": c.CRApiVersion,
		"kind": c.CRKind,
		"spec": crVal,
	}
	
	for ;crVal != NIL; {
		kubeClient.Create(ctx, CR)
		crVal = crGenerator.Next()
		CR = map[string]string{
			"apiVersion": c.CRApiVersion,
			"kind": c.CRKind,
			"spec": crVal,
		}
	}
	return nil
}
