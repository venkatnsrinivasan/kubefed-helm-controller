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

package controllers

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"kubefed-application-controller/controllers/util"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	federationv1 "kubefed-application-controller/api/v1"
)

const applicationFinalizer = "applicatio.finalizers.federation.kubefed.fulliautomatix.site"

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Config *rest.Config
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=federation.kubefed.fulliautomatix.site,resources=applications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=federation.kubefed.fulliautomatix.site,resources=applications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=types.kubefed.io,resources=federateddeployments;federatedservices;federatedconfigmaps,verbs=get;list;watch;create;update;patch;delete

func (r *ApplicationReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, reterr error) {
	context := context.Background()
	log := r.Log.WithValues("application", req.NamespacedName)
	var application federationv1.Application
	err := r.Get(context, req.NamespacedName, &application)

	if application.ObjectMeta.DeletionTimestamp.IsZero() {
		// Register our finalizer so that the hook is called before the application is deleted
		if !containsString(application.ObjectMeta.Finalizers, applicationFinalizer) {
			application.ObjectMeta.Finalizers = append(application.ObjectMeta.Finalizers, applicationFinalizer)
			if err := r.Update(context, &application); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// TODO: Need to clean up all the federated resources
	}
	if err != nil {
		log.Error(err, "Unable to fetch application")
		// Skip if not found
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	defer func() {
		err := r.Client.Update(context, &application)
		if err != nil {
			log.Error(err, "Unable to update status ")
			if reterr != nil {
				reterr = err
			}
		}
	}()
	application.Status.State = federationv1.Deploying
	// First validate the input application
	err = r.validateApplication(application)
	if err != nil {
		log.Error(err, "Unable to validate application")
		application.Status.State = federationv1.Errored

		// Skip if not found
		return ctrl.Result{}, err
	}
	err = r.deployApplication(application, log)
	if err != nil {
		log.Error(err, "Unable to deploy application")
		application.Status.State = federationv1.Errored

		// Skip if not found
		return ctrl.Result{}, err
	}
	application.Status.State = federationv1.Deployed

	application.Status.DeployedTimestamp = &metav1.Time{Time: time.Now()}
	return ctrl.Result{}, nil
}
func (r *ApplicationReconciler) validateApplication(application federationv1.Application) error {
	if application.Spec.Type == "" || application.Spec.Type != federationv1.Helm {
		return fmt.Errorf("Invalid application type %s .Only Helm is supported", application.Spec.Type)
	}
	chartName := application.Spec.Template.Chart.Name

	if chartName == "" {
		return fmt.Errorf("Invalid chart name %s ", chartName)
	}
	// TODO : all validation to check if its a valid chart by downloading
	return nil
}

func (r *ApplicationReconciler) deployApplication(application federationv1.Application, log logr.Logger) error {
	chartName := application.Spec.Template.Chart.Name
	chartRepo := application.Spec.Template.Chart.Repo
	helmClient, err := util.NewHelmClient(r.Config)
	if err != nil {
		return fmt.Errorf("Unable to create helm client")
	}
	template, err := helmClient.Template(application.ObjectMeta.Name, chartName, chartRepo, util.GlobalOptions{Namespace: application.Spec.Template.Chart.Namespace})
	if err != nil {
		log.Error(err, "Unable to create template for application")
		return fmt.Errorf("Unable to generate a helm template from chart %s", chartName)
	}

	kubefedConverter, err := util.NewFederatedResourceConverter(template)
	if err != nil {
		return fmt.Errorf("Unable to create a kubefedctl converter")
	}
	fedResources, err := kubefedConverter.GenerateFederatedUnstructuredList(template)
	if err != nil {
		return fmt.Errorf("Unable to  generate a federated manifest")
	}
	dynamicClient, err := util.NewServerSideDeployer(r.Config)
	if err != nil {
		return fmt.Errorf("Unable to create a dynamic client")
	}
	for _, eachFederatedResource := range fedResources {
		err = dynamicClient.Apply(*eachFederatedResource, application.Spec.Template.Chart.Namespace)
	}

	//TODO add support for specific chart versions.
	return err
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&federationv1.Application{}).
		Complete(r)
}
