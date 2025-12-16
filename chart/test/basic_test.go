package main

import (
	"context"
	"fmt"
	"net"
	nethttp "net/http"
	"net/url"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons-test/helm"
	"github.com/flanksource/commons/http"
	"github.com/google/uuid"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mission Control", ginkgo.Ordered, func() {
	var mcStopChan, configDBStopChan chan struct{}

	BeforeAll(func() {
		By("Installing Mission Control")
		mcChart = helm.NewHelmChart(ctx, "../")
		// mcChart = helm.NewHelmChart(ctx, "flanksource/mission-control")

		/*
			Expect(mcChart.
				Release("mission-control").Namespace("mission-control").
			WaitFor(time.Minute * 15).
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
						"logLevel": "-vvv",
					},
					"logLevel": "-vvv",
				}).
				InstallOrUpgrade()).NotTo(HaveOccurred())

		*/
		adminPasswordSecret, err := k8s.CoreV1().Secrets(namespace).Get(context.TODO(), "mission-control-admin-password", v1.GetOptions{})
		Expect(err).NotTo(HaveOccurred(), "Failed to get Mission Control admin password secret")
		adminPassword := string(adminPasswordSecret.Data["password"])
		Expect(adminPassword).NotTo(BeEmpty(), "Mission Control admin password should not be empty")
		logger.Infof(clicky.MustFormat(adminPassword))

		// Port forward to mission-control pod
		var mcLocalPort int
		mcLocalPort, mcStopChan, err = portForwardPod(ctx, namespace, "app.kubernetes.io/name=mission-control", 8080)
		Expect(err).NotTo(HaveOccurred(), "Failed to port forward to Mission Control pod")

		// Initialize Mission Control client using port-forwarded address
		mcInstance = &MissionControl{
			Client:   k8s,
			HTTP:     http.NewClient().BaseURL(fmt.Sprintf("http://localhost:%d", mcLocalPort)).Auth("admin@local", adminPassword),
			Username: "admin@local",
			Password: adminPassword,
		}

		// Port forward to config-db pod
		var configDBLocalPort int
		configDBLocalPort, configDBStopChan, err = portForwardPod(ctx, namespace, "app.kubernetes.io/name=config-db", 8080)
		Expect(err).NotTo(HaveOccurred(), "Failed to port forward to Config DB pod")

		mcInstance.ConfigDB = http.NewClient().BaseURL(fmt.Sprintf("http://localhost:%d", configDBLocalPort)).Auth("admin@local", adminPassword)
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
	})

	It("Should run system scraper", func() {
		scraper := mcInstance.GetScraper(uuid.Nil.String())
		sr, err := scraper.Run()
		Expect(err).NotTo(HaveOccurred())
		logger.Infof("Scrape result is %+v", sr)
	})

	// System scraper runs and should populate job histories/local agent
	It("Should search catalog", func() {
		results, err := mcInstance.SearchCatalog("type=*")
		logger.Infof("%=v", results)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(results)).To(BeNumerically(">", 1))
	})
})

// portForwardPod sets up port forwarding to a pod matching the given label selector.
// Returns the local port, a stop channel to close when done, and any error.
func portForwardPod(ctx context.Context, namespace, labelSelector string, remotePort int) (int, chan struct{}, error) {
	// Find pod matching the label selector
	pods, err := k8s.CoreV1().Pods(namespace).List(ctx, v1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return 0, nil, fmt.Errorf("failed to list pods: %w", err)
	}
	if len(pods.Items) == 0 {
		return 0, nil, fmt.Errorf("no pods found matching selector %s", labelSelector)
	}
	podName := pods.Items[0].Name

	// Get a free local port
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, nil, fmt.Errorf("failed to get free port: %w", err)
	}
	localPort := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// Build rest config from kubeconfig
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to build rest config: %w", err)
	}

	// Build the port-forward URL
	serverURL, err := url.Parse(restConfig.Host)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to parse server URL: %w", err)
	}
	serverURL.Path = fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", namespace, podName)

	// Create SPDY transport
	transport, upgrader, err := spdy.RoundTripperFor(restConfig)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create round tripper: %w", err)
	}

	dialer := spdy.NewDialer(upgrader, &nethttp.Client{Transport: transport}, nethttp.MethodPost, serverURL)

	// Set up channels
	stopChan := make(chan struct{}, 1)
	readyChan := make(chan struct{})

	// Create port forwarder
	ports := []string{fmt.Sprintf("%d:%d", localPort, remotePort)}
	pf, err := portforward.New(dialer, ports, stopChan, readyChan, nil, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create port forwarder: %w", err)
	}

	// Start port forwarding in background
	errChan := make(chan error, 1)
	go func() {
		errChan <- pf.ForwardPorts()
	}()

	// Wait for port forward to be ready or error
	select {
	case <-readyChan:
		return localPort, stopChan, nil
	case err := <-errChan:
		return 0, nil, fmt.Errorf("port forward failed: %w", err)
	case <-ctx.Done():
		close(stopChan)
		return 0, nil, ctx.Err()
	}
}
