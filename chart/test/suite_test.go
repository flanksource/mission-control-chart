package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	flanksourceCtx "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/kubernetes"
	"github.com/flanksource/commons-test/helm"
	commonsLogger "github.com/flanksource/commons/logger"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	kubeconfig  string
	namespace   string
	chartPath   string
	releaseName string
	ctx         flanksourceCtx.Context
)

var logger commonsLogger.Logger
var k8s *kubernetes.Client

func findParentDir(dir string) string {
	currentDir, _ := os.Getwd()

	for {

		if _, ok := os.Stat(filepath.Join(currentDir, dir)); ok == nil {
			return filepath.Join(currentDir, dir)
		}
		if _, ok := os.Stat(filepath.Join(currentDir, ".git")); ok == nil {
			// Reached the git root, stop searching
			return currentDir
		}
		currentDir = filepath.Dir(currentDir)
	}
}

func TestHelm(t *testing.T) {
	logger = commonsLogger.NewWithWriter(GinkgoWriter)
	commonsLogger.Use(GinkgoWriter)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mission Control Helm Chart Suite")
}

var mcInstance *MissionControl
var mcInstanceWithoutAuth *MissionControl
var mcChart *helm.HelmChart
var _ = BeforeSuite(func() {

	// Get environment variables or use defaults
	kubeconfig = os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home := os.Getenv("HOME")
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	namespace = os.Getenv("TEST_NAMESPACE")
	if namespace == "" {
		namespace = "mission-control"
	}

	chartPath = findParentDir("chart")

	releaseName = "mission-control"

	logger.Infof("KUBECONFIG=%s ns=%s, chart=%s", kubeconfig, namespace, chartPath)

	if stat, err := os.Stat(kubeconfig); err != nil || stat.IsDir() {
		path, _ := filepath.Abs(kubeconfig)
		Skip(fmt.Sprintf("KUBECONFIG %s is not valid, skipping helm tests", path))
	}

	ctx = flanksourceCtx.New().
		WithNamespace("mission-control")

	var err error
	k8s, err = ctx.LocalKubernetes(kubeconfig)
	Expect(err).NotTo(HaveOccurred())

})
