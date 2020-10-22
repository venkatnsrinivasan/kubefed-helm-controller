/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ApplicationType string

// +kubebuilder:validation:Enum=Deploying;Errored;Deployed;Rejected
type ApplicationDeploymentState string

const (
	Deploying ApplicationDeploymentState = "Deploying"
	Deployed  ApplicationDeploymentState = "Deployed"
	Rejected  ApplicationDeploymentState = "Rejected"
	Errored   ApplicationDeploymentState = "Errored"
)
const (
	Helm ApplicationType = "Helm"
)

type ApplicationTemplateSpec struct {
	Chart HelmChartSpec `json:"chart"`
}
type HelmChartSpec struct {

	// Name of the helm chart
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace where the chart artifacts should be deployed
	// +kubebuilder:validation:Optional
	Namespace string `json:"namespace"`

	// Repository to fetch the helm chart from
	// +kubebuilder:validation:required
	Repo string `json:"repoUrl"`

	// Installing a specific version
	// +kubebuilder:validation:Optional
	Version string `json:"version,omitempty"`
}

// ApplicationSpec defines the desired state of Application
type ApplicationSpec struct {
	// Defines an application type , by default it is Helm .
	// +kubebuilder:validation:Required
	Type ApplicationType `json:"type"`
	// +kubebuilder:validation:Required
	Template ApplicationTemplateSpec `json:"template"`
}

// ApplicationStatus defines the observed state of Application
type ApplicationStatus struct {
	State ApplicationDeploymentState `json:"state,omitempty"`

	DeployedTimestamp *metav1.Time `json:"deployedAt,omitempty"`
}

// +kubebuilder:object:root=true

// Application is the Schema for the applications API
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec,omitempty"`
	Status ApplicationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApplicationList contains a list of Application
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Application{}, &ApplicationList{})
}
