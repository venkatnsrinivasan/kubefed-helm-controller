package util

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/kubefed/pkg/kubefedctl/federate"
	"sigs.k8s.io/yaml"
)

//FederatedResourceConverter
type FederatedResourceConverter interface {
	GenerateFederatedManifest(input *string) (*string, error)
	GenerateFederatedUnstructuredList(input *string) ([]*unstructured.Unstructured, error)
}

// FederatedResource impl
type FederatedResource struct {
	ResourceManifest string
	OutputYAML       bool
}

//NewFederatedResourceConverter creates a new converter instance
func NewFederatedResourceConverter(inputManifestString *string) (*FederatedResource, error) {
	return &FederatedResource{ResourceManifest: *inputManifestString, OutputYAML: true}, nil
}

func (federatedResource *FederatedResource) GenerateFederatedUnstructuredList(input *string) ([]*unstructured.Unstructured, error) {
	return federatedResource.convertToUnstructuredList(input)
}

func (federatedResource *FederatedResource) convertToUnstructuredList(input *string) ([]*unstructured.Unstructured, error) {
	resources, err := parseInputResources(input)
	if err != nil {
		return nil, err
	}
	fedresources, err := federate.FederateResources(resources)
	if err != nil {
		return nil, err
	}
	return fedresources, nil
}

//GenerateFederatedManifest outputs a federated manifest
func (federatedResource *FederatedResource) GenerateFederatedManifest(input *string) (*string, error) {

	fedresources, err := federatedResource.convertToUnstructuredList(input)
	if err != nil {
		return nil, err
	}
	var buf *bytes.Buffer = new(bytes.Buffer)
	err = federate.WriteUnstructuredObjsToYaml(fedresources, buf)
	if err != nil {
		return nil, err
	}
	result := buf.String()
	return &result, nil
}

func parseInputResources(input *string) ([]*unstructured.Unstructured, error) {
	var unstructuredList []*unstructured.Unstructured
	reader := utilyaml.NewYAMLReader(bufio.NewReader(strings.NewReader(*input)))
	for {
		unstructuedObj := &unstructured.Unstructured{}
		// Read one YAML document at a time, until io.EOF is returned
		buf, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		if len(buf) == 0 {
			break
		}
		if err := yaml.Unmarshal(buf, unstructuedObj); err != nil {
			return nil, err
		}
		unstructuredList = append(unstructuredList, unstructuedObj)
	}

	return unstructuredList, nil
}
