package util

import (
	"fmt"
	"log"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
)

// HelmClient interface
type HelmClient interface {
	Template(releaseName string, chartName string, chartRepo string, options GlobalOptions) (*string, error)
}

// Helm properties struct
type Helm struct {
	kubeconfig *rest.Config
}
type GlobalOptions struct {
	Namespace string
}

// NewHelmClient creates and intializes a helmclient
func NewHelmClient(kubeconfig *rest.Config) (HelmClient, error) {
	helm := Helm{}
	helm.kubeconfig = kubeconfig
	return &helm, nil
}

func (helm *Helm) Template(releaseName string, chartName string, chartRepo string, options GlobalOptions) (*string, error) {
	config, err := helm.createConfig(options)
	if err != nil {
		return nil, err
	}

	installer := action.NewInstall(config)
	installer.DryRun = true
	installer.ClientOnly = true
	installer.ReleaseName = releaseName
	installer.Namespace = options.Namespace
	installer.RepoURL = chartRepo
	//configFlags := helm.getConfigFlags(options.namespace)

	settings := cli.New()
	chartPath, err := installer.ChartPathOptions.LocateChart(chartName, settings)
	if err != nil {
		return nil, err
	}
	//TODO merge values
	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, err
	}

	p := getter.All(settings)
	valueOpts := values.Options{}
	vals, err := valueOpts.MergeValues(p)
	if err != nil {
		return nil, err
	}
	rel, err := installer.Run(chart, vals)
	if err != nil {
		return nil, err
	}
	return &rel.Manifest, nil

}

func (helm *Helm) createConfig(options GlobalOptions) (*action.Configuration, error) {
	configFlags := helm.getConfigFlags(options.Namespace)
	actionConfig := new(action.Configuration)
	err := actionConfig.Init(configFlags, options.Namespace, "", debugLog)
	return actionConfig, err
}
func (helm *Helm) getConfigFlags(namespace string) *genericclioptions.ConfigFlags {
	return &genericclioptions.ConfigFlags{
		Namespace:   &namespace,
		APIServer:   &helm.kubeconfig.Host,
		CAFile:      &helm.kubeconfig.CAFile,
		BearerToken: &helm.kubeconfig.BearerToken,
	}
}

func debugLog(format string, v ...interface{}) {
	format = fmt.Sprintf("[debug] %s\n", format)
	log.Output(2, fmt.Sprintf(format, v...))
}

func DerefString(s *string) string {
	if s != nil {
		return *s
	}

	return ""
}
