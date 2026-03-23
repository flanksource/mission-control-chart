package main

import (
	"context"
	"strings"
	"time"

	"github.com/flanksource/commons-test/helm"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Mission Control (Kratos)", ginkgo.Ordered, Label("kratos"), func() {
	const (
		kratosNamespace   = "mission-control-kratos-test"
		kratosReleaseName = "mission-control-kratos"
	)

	BeforeAll(func() {
		By("Installing Mission Control with Kratos auth")

		host := "mission-control-kratos.cluster.local"
		Expect(helm.NewHelmChart(ctx, "../").
			Release(kratosReleaseName).Namespace(kratosNamespace).
			WaitFor(time.Minute * 7).
			Values(map[string]any{
				"authProvider": "kratos",
				"global": map[string]any{
					"ui": map[string]any{
						"host": host,
					},
				},
				// Keep this close to defaults; disable ingress/UI only for CI isolation.
				"ingress": map[string]any{"enabled": false},
				"flanksource-ui": map[string]any{"enabled": false},
			}).
			InstallOrUpgrade()).NotTo(HaveOccurred())
	})

	It("Should render mission-control kratos config from default template", func() {
		cm, err := k8s.CoreV1().ConfigMaps(kratosNamespace).Get(context.TODO(), "mission-control-kratos-config", v1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		kratosYAML, ok := cm.Data["kratos.yaml"]
		Expect(ok).To(BeTrue())
		Expect(kratosYAML).NotTo(BeEmpty())

		// These values come from chart/files/kratos-config.yaml and should stay in sync.
		Expect(kratosYAML).To(ContainSubstring("base_url: https://mission-control-kratos.cluster.local/api/.ory"))
		Expect(kratosYAML).To(ContainSubstring("default_browser_return_url: https://mission-control-kratos.cluster.local/"))
		Expect(kratosYAML).To(ContainSubstring("allowed_return_urls:"))
		Expect(kratosYAML).To(ContainSubstring("- https://mission-control-kratos.cluster.local"))
		Expect(kratosYAML).To(ContainSubstring("identity:"))
		Expect(kratosYAML).To(ContainSubstring("url: base64://"))
		Expect(strings.Count(kratosYAML, "mission-control-kratos.cluster.local")).To(BeNumerically(">", 3))
	})

	It("Should run kratos automigration and start kratos pod", func() {
		Eventually(func(g Gomega) {
			pods, err := k8s.CoreV1().Pods(kratosNamespace).List(context.TODO(), v1.ListOptions{
				LabelSelector: "app.kubernetes.io/name=kratos",
			})
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(pods.Items).NotTo(BeEmpty())

			pod := pods.Items[0]
			g.Expect(pod.Status.Phase).To(Equal("Running"))

			initialized := false
			for _, c := range pod.Status.Conditions {
				if c.Type == "Initialized" && c.Status == "True" {
					initialized = true
					break
				}
			}
			g.Expect(initialized).To(BeTrue())

			hasAutomigrate := false
			for _, init := range pod.Status.InitContainerStatuses {
				if init.Name == "kratos-automigrate" {
					hasAutomigrate = true
					g.Expect(init.State.Terminated).NotTo(BeNil())
					g.Expect(init.State.Terminated.ExitCode).To(Equal(int32(0)))
				}
			}
			g.Expect(hasAutomigrate).To(BeTrue())

			ready := false
			for _, c := range pod.Status.ContainerStatuses {
				if c.Name == "kratos" {
					ready = c.Ready
					break
				}
			}
			g.Expect(ready).To(BeTrue())
		}).WithTimeout(4 * time.Minute).WithPolling(5 * time.Second).Should(Succeed())
	})
})
