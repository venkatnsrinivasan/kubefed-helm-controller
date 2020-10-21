package util

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

type DynamicClient interface {
	Apply(resourceObj unstructured.Unstructured, namespace string) error
}

type ServerSideDeployer struct {
	config        *rest.Config
	dynamicClient dynamic.Interface
	restMapper    *restmapper.DeferredDiscoveryRESTMapper
}

func NewServerSideDeployer(config *rest.Config) (*ServerSideDeployer, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(discoveryClient))

	// 2. Prepare the dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &ServerSideDeployer{config: config, dynamicClient: dynamicClient, restMapper: mapper}, nil
}

func (ssd *ServerSideDeployer) Apply(resourceObj unstructured.Unstructured, namespace string) error {
	// first get the gvk
	gvk := resourceObj.GroupVersionKind()
	// find the group version resource for the gvk
	gvrMapping, err := ssd.restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return err
	}

	dynamicResource := ssd.dynamicClient.Resource(gvrMapping.Resource).Namespace(namespace)
	resourceObjJson, err := runtime.Encode(unstructured.UnstructuredJSONScheme, &resourceObj)
	if err != nil {
		return err
	}

	_, err = dynamicResource.Patch(resourceObj.GetName(), types.ApplyPatchType, resourceObjJson, metav1.PatchOptions{
		FieldManager: "kubefed-helm-controller",
	})

	return err
}
