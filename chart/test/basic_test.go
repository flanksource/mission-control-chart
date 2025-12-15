package main

import (
	"context"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons-test/helm"
	"github.com/flanksource/commons/http"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Mission Control", func() {
	It("Basic Auth", func() {
		By("Installing Mission Control")
		mcChart = helm.NewHelmChart(ctx, "../")
		// mcChart = helm.NewHelmChart(ctx, "flanksource/mission-control")
		Expect(mcChart.
			Release("mission-control").
			Namespace("mission-control").
			Wait().
			Values(map[string]interface{}{
				"global": map[string]interface{}{
					"ui": map[string]interface{}{
						"host": "mission-control.cluster.local",
					},
				},
				"authProvider": "basic",

				"htpasswd": map[string]interface{}{
					"create": true,
				},
				"ingress": map[string]interface{}{
					"enabled": true,
					"annotations": map[string]interface{}{
						"cert-manager.io/cluster-issuer": "self-signed",
					},
				},
				"config-db": map[string]interface{}{
					"enabled":  false,
					"logLevel": "-vvv",
				},
				"logLevel": "-vvv",
			}).
			InstallOrUpgrade()).NotTo(HaveOccurred())

		adminPasswordSecret, err := k8s.CoreV1().Secrets(namespace).Get(context.TODO(), "mission-control-admin-password", v1.GetOptions{})
		Expect(err).NotTo(HaveOccurred(), "Failed to get Mission Control admin password secret")
		adminPassword := string(adminPasswordSecret.Data["password"])
		Expect(adminPassword).NotTo(BeEmpty(), "Mission Control admin password should not be empty")
		logger.Infof(clicky.MustFormat(adminPassword))

		podIP, err := k8s.GetPodIP(ctx, namespace, "app.kubernetes.io/name=mission-control")
		Expect(err).NotTo(HaveOccurred(), "Failed to get Mission Control pod IP")

		// Initialize Mission Control client, get credentianls and serviceIP from deployed chart.
		mcInstance = &MissionControl{
			Client:   k8s,
			HTTP:     http.NewClient().BaseURL("http://"+podIP+":8080").Auth("admin@local", adminPassword),
			Username: "admin@local",
			Password: adminPassword,
		}
		podIP, err = k8s.GetPodIP(ctx, namespace, "app.kubernetes.io/name=config-db")
		Expect(err).NotTo(HaveOccurred(), "Failed to get Config DB pod IPs")
		mcInstance.ConfigDB = http.NewClient().BaseURL("http://"+podIP+":8080").Auth("admin@local", adminPassword)
	})

	It("SearchCatalog", func() {
		results, err := mcInstance.SearchCatalog("type=*")
		logger.Infof("%=v", results)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(results)).To(BeNumerically(">", 1))
	})
})
