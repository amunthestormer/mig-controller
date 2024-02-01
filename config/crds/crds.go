package crds

import (
	apiextinstall "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/install"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"
	"os"
)

var crdYamlFiles = []string{
	"migration.openshift.io_directimagemigrations.yaml",
	"migration.openshift.io_directimagestreammigrations.yaml",
	"migration.openshift.io_directvolumemigrationprogresses.yaml",
	"migration.openshift.io_directvolumemigrations.yaml",
	"migration.openshift.io_miganalytics.yaml",
	"migration.openshift.io_migclusters.yaml",
	"migration.openshift.io_migmigrations.yaml",
	"migration.openshift.io_mighooks.yaml",
	"migration.openshift.io_migstorages.yaml",
	"migration.openshift.io_migplans.yaml",
}

func Crds() ([]*apiextv1.CustomResourceDefinition, error) {
	apiextinstall.Install(scheme.Scheme)
	rawCRDs, err := createCRDs()
	if err != nil {
		return nil, err
	}
	var objs []*apiextv1.CustomResourceDefinition
	for _, rawCRD := range rawCRDs {
		crdObj := &apiextv1.CustomResourceDefinition{}
		err = yaml.Unmarshal(rawCRD, crdObj)
		if err != nil {
			return nil, err
		}
		objs = append(objs, crdObj)
	}
	return objs, nil
}
func createCRDs() ([][]byte, error) {
	crdBytesList := make([][]byte, 0)
	crdBytes := make([]byte, 0)
	var err error
	for _, crdYamlFile := range crdYamlFiles {
		crdBytes, err = os.ReadFile(crdYamlFile)
		if err != nil {
			return nil, err
		}
		crdBytesList = append(crdBytesList, crdBytes)
	}
	return crdBytesList, nil
}
