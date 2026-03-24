package main

import (
	"bytes"
	"context"
	"fmt"
	"net"
	nethttp "net/http"
	"net/url"
	"os/exec"
	"time"

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

	return portForwardResource(ctx, namespace, fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", namespace, podName), remotePort)
}

// portForwardService sets up port forwarding to a service using kubectl.
// Returns the local port, a stop channel to close when done, and any error.
func portForwardService(ctx context.Context, namespace, serviceName string, remotePort int) (int, chan struct{}, error) {
	// Get a free local port
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, nil, fmt.Errorf("failed to get free port: %w", err)
	}
	localPort := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	var output bytes.Buffer
	cmd := exec.CommandContext(
		ctx,
		"kubectl",
		"--kubeconfig", kubeconfig,
		"-n", namespace,
		"port-forward",
		"svc/"+serviceName,
		fmt.Sprintf("%d:%d", localPort, remotePort),
	)
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Start(); err != nil {
		return 0, nil, fmt.Errorf("failed to start kubectl port-forward for service %s: %w", serviceName, err)
	}

	stopChan := make(chan struct{}, 1)
	errChan := make(chan error, 1)

	go func() {
		errChan <- cmd.Wait()
	}()

	go func() {
		<-stopChan
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}()

	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		conn, dialErr := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", localPort), 500*time.Millisecond)
		if dialErr == nil {
			_ = conn.Close()
			return localPort, stopChan, nil
		}

		select {
		case waitErr := <-errChan:
			return 0, nil, fmt.Errorf("kubectl port-forward for service %s exited early: %w\noutput: %s", serviceName, waitErr, output.String())
		case <-ctx.Done():
			close(stopChan)
			return 0, nil, ctx.Err()
		default:
		}

		time.Sleep(200 * time.Millisecond)
	}

	close(stopChan)
	return 0, nil, fmt.Errorf("timed out waiting for kubectl port-forward to service %s\noutput: %s", serviceName, output.String())
}

func portForwardResource(ctx context.Context, namespace, apiPath string, remotePort int) (int, chan struct{}, error) {
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
	serverURL.Path = apiPath

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
