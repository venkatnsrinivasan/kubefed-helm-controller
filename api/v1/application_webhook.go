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
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var applicationlog = logf.Log.WithName("application-resource")

func (r *Application) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-federation-kubefed-fulliautomatix-site-v1-application,mutating=true,failurePolicy=fail,groups=federation.kubefed.fulliautomatix.site,resources=applications,verbs=create;update,versions=v1,name=mapplication.kb.io

var _ webhook.Defaulter = &Application{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Application) Default() {
	applicationlog.Info("default", "name", r.Name)

	if r.Spec.Type == "" {
		r.Spec.Type = Helm
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-federation-kubefed-fulliautomatix-site-v1-application,mutating=false,failurePolicy=fail,groups=federation.kubefed.fulliautomatix.site,resources=applications,versions=v1,name=vapplication.kb.io

var _ webhook.Validator = &Application{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Application) ValidateCreate() error {
	applicationlog.Info("validate create", "name", r.Name)
	return r.validateApplication()
}

func (application *Application) validateApplication() error {
	if application.Spec.Type == "" || application.Spec.Type != Helm {
		return fmt.Errorf("Invalid application type %s .Only Helm is supported", application.Spec.Type)
	}
	chartName := application.Spec.Template.Chart.Name

	if chartName == "" {
		return fmt.Errorf("Invalid/empty chart name")
	}
	repourl := application.Spec.Template.Chart.Repo

	if repourl == "" {
		return fmt.Errorf("Repo url is a required field ")
	}
	// TODO : maybe add validation to check if its a valid chart by downloading
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Application) ValidateUpdate(old runtime.Object) error {
	applicationlog.Info("validate update", "name", r.Name)
	return r.validateApplication()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Application) ValidateDelete() error {
	applicationlog.Info("validate delete", "name", r.Name)

	return nil
}
