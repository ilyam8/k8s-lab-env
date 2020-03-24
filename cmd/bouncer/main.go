package main

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s-lab-env/pkg/bouncer"
	"k8s-lab-env/pkg/clientset"
	"k8s.io/client-go/kubernetes"
)

const namespace = "netdata-auto-discovery"

func isInCluster() bool {
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	return len(host) != 0 && len(port) != 0
}

func newClientSet() (*kubernetes.Clientset, error) {
	if isInCluster() {
		return clientset.InCluster()
	}
	return clientset.OutOfCluster()
}

func main() {
	log.SetFormatter(&log.TextFormatter{TimestampFormat: "15:04:05"})
	c, err := newClientSet()
	if err != nil {
		log.Fatal(err)
	}

	b := bouncer.Bouncer{
		LabelSelector: "",
		FieldSelector: "",
		RetryTimeout:  5 * time.Second,
		BounceEvery:   30 * time.Second,
		Client:        c.AppsV1().Deployments(namespace),
	}

	b.Bounce()
}
