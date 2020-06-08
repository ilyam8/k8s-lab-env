package main

import (
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s-lab-env/pkg/bouncer"
	"k8s-lab-env/pkg/clientset"
	"k8s.io/client-go/kubernetes"
)

const namespace = "netdata-service-discovery"

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

func bounceEvery() time.Duration {
	v, ok := os.LookupEnv("BOUNCE_EVERY")
	if !ok {
		return 30
	}
	be, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 30
	}
	return time.Duration(be)
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
		BounceEvery:   bounceEvery() * time.Second,
		Client:        c.AppsV1().Deployments(namespace),
	}

	b.Bounce()
}
