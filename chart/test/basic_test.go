package main

import (
	"context"
	"fmt"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons-test/helm"
	"github.com/flanksource/commons/http"
	"github.com/google/uuid"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mission Control", ginkgo.Ordered, func() {
	var mcStopChan, configDBStopChan chan struct{}

	BeforeAll(func() {
		By("Installing Mission Control")

		Expect(helm.NewHelmChart(ctx, "../").
			Release("mission-control").Namespace("mission-control").
			WaitFor(time.Minute * 5).
			Values(map[string]any{
				"global": map[string]any{
					"ui": map[string]any{
						"host": "mission-control.cluster.local",
					},
				},
				"authProvider": "basic",
				"htpasswd": map[string]any{
					"create": true,
				},
				"kratos": map[string]any{
					"enabled": false,
				},
				"ingress": map[string]any{
					"enabled": false,
				},
				"config-db": map[string]any{
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

		// Port forward to mission-control pod
		var mcLocalPort int
		mcLocalPort, mcStopChan, err = portForwardPod(ctx, namespace, "app.kubernetes.io/name=mission-control", 8080)
		Expect(err).NotTo(HaveOccurred(), "Failed to port forward to Mission Control pod")

		// Port forward to config-db pod
		var configDBLocalPort int
		configDBLocalPort, configDBStopChan, err = portForwardPod(ctx, namespace, "app.kubernetes.io/name=config-db", 8080)
		Expect(err).NotTo(HaveOccurred(), "Failed to port forward to Config DB pod")

		// Initialize Mission Control client using port-forwarded address
		mcInstance = &MissionControl{
			Client:   k8s,
			HTTP:     http.NewClient().BaseURL(fmt.Sprintf("http://localhost:%d", mcLocalPort)).Auth("admin@local", adminPassword),
			Username: "admin@local",
			Password: adminPassword,
			ConfigDB: http.NewClient().BaseURL(fmt.Sprintf("http://localhost:%d", configDBLocalPort)).Auth("admin@local", adminPassword),
		}

		mcInstanceWithoutAuth = &MissionControl{
			Client:   k8s,
			HTTP:     http.NewClient().BaseURL(fmt.Sprintf("http://localhost:%d", mcLocalPort)),
			ConfigDB: http.NewClient().BaseURL(fmt.Sprintf("http://localhost:%d", configDBLocalPort)).Auth("admin@local", adminPassword),
		}

	})

	AfterAll(func() {
		if mcStopChan != nil {
			close(mcStopChan)
		}
		if configDBStopChan != nil {
			close(configDBStopChan)
		}
	})

	It("Should be healthy", func() {
		healthy, err := mcInstance.IsHealthy()
		Expect(err).NotTo(HaveOccurred(), "Unable to query health endpoint")
		Expect(healthy).To(BeTrue())

		healthy, err = mcInstanceWithoutAuth.IsHealthy()
		Expect(err).NotTo(HaveOccurred(), "Unable to query health endpoint")
		Expect(healthy).To(BeTrue())
	})

	It("Should run WhoAmI", func() {
		whoami, ok, err := mcInstance.WhoAmI()
		Expect(err).NotTo(HaveOccurred(), "Unable to query whoami endpoint")
		Expect(ok).To(BeTrue())
		Expect(whoami["message"]).To(Equal("success"))

		whoami, ok, err = mcInstanceWithoutAuth.WhoAmI()
		Expect(err).NotTo(HaveOccurred(), "Unable to query whoami endpoint")
		Expect(ok).To(BeFalse())
		Expect(whoami["message"]).To(Equal("Unauthorized"))
	})

	It("Should run system scraper", func() {
		scraper := mcInstance.GetScraper(uuid.Nil.String())
		sr, err := scraper.Run()
		Expect(err).NotTo(HaveOccurred())
		logger.Infof("Scrape Result:\n%s", clicky.MustFormat(sr))

		scraper = mcInstanceWithoutAuth.GetScraper(uuid.Nil.String())
		sr, err = scraper.Run()
		Expect(err).NotTo(HaveOccurred(), "basic auth")
	})

	// System scraper runs and should populate job histories/local agent
	It("Should search catalog", func() {
		results, err := mcInstance.SearchCatalog("type=*")
		logger.Infof("%s", clicky.MustFormat(results))
		Expect(err).NotTo(HaveOccurred())
		Expect(len(results)).To(BeNumerically(">", 1))
	})

	It("Should install default views", func() {
		expectedViews := []string{
			"mission-control-dashboard",
			"mission-control-system",
			"jobhistory",
		}

		Eventually(func(g Gomega) {
			views, err := k8s.List(context.TODO(), "View", namespace, "")
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(views).NotTo(BeEmpty())

			viewNames := make(map[string]struct{}, len(views))
			for _, view := range views {
				viewNames[view.GetName()] = struct{}{}
			}

			for _, name := range expectedViews {
				g.Expect(viewNames).To(HaveKey(name))
			}
		}).WithTimeout(2 * time.Minute).WithPolling(5 * time.Second).Should(Succeed())
	})
})
