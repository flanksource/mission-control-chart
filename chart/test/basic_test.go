package main

import (
	"context"
	"fmt"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons-test/helm"
	"github.com/flanksource/commons/http"
	"github.com/google/uuid"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Mission Control - Basic", ginkgo.Ordered, Label("basic"), func() {
	var mcStopChan, configDBStopChan chan struct{}

	BeforeAll(func() {
		By("Installing Mission Control")

		Expect(helm.NewHelmChart(ctx, "../").
			Release("mission-control").Namespace("mission-control").
			ForceConflicts().
			Values(map[string]any{
				"global": map[string]any{
					"ui": map[string]any{
						"host": "mission-control.cluster.local",
					},
				},
				"artifactstore": map[string]any{
					"enabled": true,
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

		Expect(waitForPodReady(ctx, namespace, "app.kubernetes.io/name=mission-control", 5*time.Minute)).To(Succeed(), "mission-control pod should become ready")
		Expect(waitForPodReady(ctx, namespace, "app.kubernetes.io/name=config-db", 5*time.Minute)).To(Succeed(), "config-db pod should become ready")

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
		healthy, statusCode, body, err := mcInstance.IsHealthy()
		Expect(err).NotTo(HaveOccurred(), "Unable to query health endpoint")
		Expect(healthy).To(BeTrue(), "Expected mission-control to be healthy, got status code: %d, body: %s", statusCode, body)

		healthy, statusCode, body, err = mcInstanceWithoutAuth.IsHealthy()
		Expect(err).NotTo(HaveOccurred(), "Unable to query health endpoint")
		Expect(healthy).To(BeTrue(), "Expected mission-control (no auth) to be healthy, got status code: %d, body: %s", statusCode, body)
	})

	It("Should run WhoAmI", func() {
		whoami, ok, err := mcInstance.WhoAmI()
		Expect(err).NotTo(HaveOccurred(), "Unable to query whoami endpoint")
		Expect(ok).To(BeTrue())
		Expect(whoami["message"]).To(Equal("success"))

		whoami, ok, err = mcInstanceWithoutAuth.WhoAmI()
		Expect(err).NotTo(HaveOccurred(), "Unable to query whoami endpoint")
		Expect(ok).To(BeFalse())
		Expect(whoami["error"]).To(Equal("unauthorized"))
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

	It("Should run the rclone artifactstore pod", func() {
		Eventually(func(g Gomega) {
			pods, err := k8s.CoreV1().Pods(namespace).List(context.TODO(), v1.ListOptions{
				LabelSelector: "app.kubernetes.io/component=artifactstore",
			})
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(pods.Items).NotTo(BeEmpty())

			pod := pods.Items[0]
			g.Expect(pod.Name).To(ContainSubstring("artifactstore"))
			g.Expect(string(pod.Status.Phase)).To(Equal("Running"))

			ready := false
			for _, c := range pod.Status.ContainerStatuses {
				if c.Name == "rclone" {
					ready = c.Ready
					break
				}
			}
			g.Expect(ready).To(BeTrue())
		}).WithTimeout(4 * time.Minute).WithPolling(5 * time.Second).Should(Succeed())
	})

	It("Should test default artifactstore connection", func() {
		connection, err := k8s.Get(context.TODO(), "Connection", namespace, "default-artifactstore")
		Expect(err).NotTo(HaveOccurred(), "default-artifactstore connection should exist")

		connectionID := string(connection.GetUID())
		Expect(connectionID).NotTo(BeEmpty())

		Eventually(func(g Gomega) {
			response, err := mcInstance.POST("/connection/test/"+connectionID, nil)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(response.IsOK()).To(BeTrue(), "expected 200, got %d", response.StatusCode)

			body, err := response.AsJSON()
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(body["message"]).To(Equal("ok"), "unexpected response body: %v", body)
		}).WithTimeout(4 * time.Minute).WithPolling(5 * time.Second).Should(Succeed())
	})

	// System scraper runs and should populate job histories/local agent
	It("Should search catalog", func() {
		results, err := mcInstance.SearchCatalog("type=*")
		logger.Infof("%s", clicky.MustFormat(results))
		Expect(err).NotTo(HaveOccurred())
		Expect(len(results)).To(BeNumerically(">", 1))
	})

	Context("Views", func() {
		// Installed by this chart
		defaultViews := []string{
			"mission-control-dashboard",
			"mission-control-system",
			"jobhistory",
			"recent-changes",
			"notification-send-history",
			"unhealthy-configs",
		}

		// Comes from subchart
		kubernetesViews := []string{
			"namespace",
			"pod",
			"pods",
			"helm-release",
			"deployments",
		}

		DescribeTable("Should install views",
			func(expected []string) {
				Eventually(func(g Gomega) {
					views, err := k8s.List(context.TODO(), "View", namespace, "")
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(views).NotTo(BeEmpty())

					viewNames := make(map[string]struct{}, len(views))
					for _, view := range views {
						viewNames[view.GetName()] = struct{}{}
					}

					for _, name := range expected {
						g.Expect(viewNames).To(HaveKey(name))
					}
				}).WithTimeout(2 * time.Minute).WithPolling(5 * time.Second).Should(Succeed())
			},
			Entry("default views", defaultViews),
			Entry("kubernetes views", kubernetesViews),
		)
	})
})
