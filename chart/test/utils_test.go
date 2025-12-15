package main

import (
	"context"
	"fmt"
	"net"
	nethttp "net/http"
	"net/url"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

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
