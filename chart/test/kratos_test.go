package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/flanksource/commons-test/helm"
	"github.com/flanksource/commons/http"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func findChromeBinary() string {
	candidates := []string{"google-chrome", "chromium", "chromium-browser", "chrome"}
	for _, candidate := range candidates {
		if path, err := exec.LookPath(candidate); err == nil {
			return path
		}
	}
	return ""
}

var _ = Describe("Mission Control (Kratos)", ginkgo.Ordered, Label("kratos"), func() {
	const (
		kratosNamespace   = "mission-control-kratos-test"
		kratosReleaseName = "mission-control"
	)

	var kratosMCStopChan, uiStopChan chan struct{}
	var kratosMC *MissionControl
	var uiLocalPort int
	const (
		kratosTestIdentifier = "admin@local"
		kratosTestPassword   = "admin"
	)

	BeforeAll(func() {
		By("Installing Mission Control with Kratos auth")

		host := "mission-control-kratos.cluster.local"
		Expect(helm.NewHelmChart(ctx, "../").
			Release(kratosReleaseName).Namespace(kratosNamespace).
			WaitFor(time.Minute * 7).
			Values(map[string]any{
				"cleanupResourcesOnDelete": true,
				"authProvider":             "kratos",
				"global": map[string]any{
					"ui": map[string]any{
						"host": host,
					},
				},
				// Keep this close to defaults; disable ingress only for CI isolation.
				"ingress":        map[string]any{"enabled": false},
				"flanksource-ui": map[string]any{"enabled": true},
			}).
			InstallOrUpgrade()).NotTo(HaveOccurred())

		mcLocalPort, stopChan, err := portForwardPod(ctx, kratosNamespace, "app.kubernetes.io/name=mission-control", 8080)
		Expect(err).NotTo(HaveOccurred(), "Failed to port forward to Mission Control pod")
		kratosMCStopChan = stopChan

		kratosMC = &MissionControl{
			Client: k8s,
			HTTP:   http.NewClient().BaseURL(fmt.Sprintf("http://localhost:%d", mcLocalPort)),
		}

		uiLocalPort, stopChan, err = portForwardPod(ctx, kratosNamespace, "app.kubernetes.io/name=incident-manager-ui", 8080)
		Expect(err).NotTo(HaveOccurred(), "Failed to port forward to incident-manager-ui pod")
		uiStopChan = stopChan
	})

	AfterAll(func() {
		if kratosMCStopChan != nil {
			close(kratosMCStopChan)
		}
		if uiStopChan != nil {
			close(uiStopChan)
		}
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
			g.Expect(string(pod.Status.Phase)).To(Equal("Running"))

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

	It("Should hit whoami endpoint with UI login and succeed", func() {
		chromePath := findChromeBinary()
		if chromePath == "" {
			Skip("chrome/chromium binary not found; skipping UI login flow test")
		}

		allocatorOpts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.ExecPath(chromePath),
			chromedp.Headless,
			chromedp.NoFirstRun,
			chromedp.NoDefaultBrowserCheck,
			chromedp.Flag("no-sandbox", true),
			chromedp.Flag("disable-dev-shm-usage", true),
		)

		allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), allocatorOpts...)
		defer cancelAlloc()

		browserCtx, cancelBrowser := chromedp.NewContext(allocCtx)
		defer cancelBrowser()

		runCtx, cancelRun := context.WithTimeout(browserCtx, 2*time.Minute)
		defer cancelRun()

		loginURL := fmt.Sprintf("http://localhost:%d/login", uiLocalPort)
		identifierSelector := `input[name="identifier"], input[name="email"], input[name="traits.email"], input[type="email"], input[name="username"]`
		passwordSelector := `input[name="password"], input[name="traits.password"], input[type="password"]`
		submitSelector := `button[type="submit"], input[type="submit"]`

		err := chromedp.Run(runCtx,
			chromedp.Navigate(loginURL),
			chromedp.WaitVisible("body", chromedp.ByQuery),
			chromedp.WaitVisible(identifierSelector, chromedp.ByQuery),
			chromedp.WaitVisible(passwordSelector, chromedp.ByQuery),
			chromedp.SendKeys(identifierSelector, kratosTestIdentifier, chromedp.ByQuery),
			chromedp.SendKeys(passwordSelector, kratosTestPassword, chromedp.ByQuery),
			chromedp.Click(submitSelector, chromedp.ByQuery),
		)
		Expect(err).NotTo(HaveOccurred(), "Failed to login via UI")

		Eventually(func() bool {
			var whoamiOK bool
			err := chromedp.Run(runCtx,
				chromedp.Evaluate(`(async () => {
					const paths = ["/auth/whoami", "/api/auth/whoami"]
					for (const path of paths) {
						try {
							const res = await fetch(path, { credentials: "include" })
							if (res.ok) {
								return true
							}
						} catch (e) {
						}
					}
					return false
				})()`, &whoamiOK),
			)
			if err != nil {
				return false
			}
			return whoamiOK
		}).WithTimeout(45 * time.Second).WithPolling(2 * time.Second).Should(BeTrue())
	})

	It("Should hit whoami endpoint without session and get authorization error", func() {
		whoami, ok, err := kratosMC.WhoAmI()
		Expect(err).NotTo(HaveOccurred())
		Expect(ok).To(BeFalse())
		Expect(whoami["error"]).To(Equal("Authorization Error"))
	})
})
