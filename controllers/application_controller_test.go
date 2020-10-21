package controllers

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	appv1 "kubefed-application-controller/api/v1"
	"time"
)

var _ = Describe("application controller", func() {
	const (
		AppName      = "application-test"
		AppNameSpace = "default"
		duration     = time.Second * 10
		interval     = time.Millisecond * 250
		timeout      = time.Second * 30
	)

	Context("When creating new application ", func() {
		It("Should created federated resources ", func() {
			By("Creating a new Application", func() {
				ctx := context.Background()
				application := &appv1.Application{
					TypeMeta: metav1.TypeMeta{
						Kind:       "federation.kubefed.fulliautomatix.site/v1",
						APIVersion: "Application",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      AppName,
						Namespace: AppNameSpace,
					},
					Spec: appv1.ApplicationSpec{
						Type: "Helm",
						Template: appv1.ApplicationTemplateSpec{
							Chart: appv1.HelmChartSpec{
								Name:      "nginx",
								Namespace: "kubefed-poc",
								Repo:      "https://charts.bitnami.com/bitnami",
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, application)).Should(Succeed())
				createdApp := &appv1.Application{}
				Eventually(func() appv1.ApplicationDeploymentState {
					err := k8sClient.Get(ctx, types.NamespacedName{Name: AppName, Namespace: AppNameSpace}, createdApp)
					if err != nil {
						return ""
					}
					return createdApp.Status.State
				}, timeout, interval).Should(Equal(appv1.Deployed))

			})

		})
	})

})
