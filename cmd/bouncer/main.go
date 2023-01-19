package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s-lab-env/pkg/bouncer"
	"k8s-lab-env/pkg/clientset"
	"k8s.io/client-go/kubernetes"
)

const namespace = "netdata-service-discovery-lab"

func main() {
	log.SetFormatter(&log.TextFormatter{TimestampFormat: "15:04:05"})
	c, err := newClientSet()
	if err != nil {
		log.Fatal(err)
	}

	ns := namespace
	if v, ok := os.LookupEnv("BOUNCE_NAMESPACE"); ok {
		ns = v
	}

	b := bouncer.Bouncer{
		Client:            c.AppsV1().Deployments(ns),
		LabelSelector:     "",
		FieldSelector:     "",
		RetryTimeout:      5 * time.Second,
		BounceEvery:       time.Duration(lookupIntEnvVar("BOUNCE_EVERY", 30)) * time.Second,
		RandomBouncing:    false,
		MinReplicas:       int32(lookupIntEnvVar("MIN_REPLICAS", 1)),
		MaxReplicas:       int32(lookupIntEnvVar("MAX_REPLICAS", 10)),
		MaxBounceReplicas: int32(lookupIntEnvVar("MAX_BOUNCE_REPLICAS", 1)),
		DryRun:            false,
	}

	fmt.Printf("ns=%s, bounce_every=%s, random=%v, min_repl=%d, max_repl=%d, max_bounce_repl=%d\n",
		ns, b.BounceEvery, b.RandomBouncing, b.MinReplicas, b.MaxReplicas, b.MaxBounceReplicas)

	b.Bounce()
}

func newClientSet() (*kubernetes.Clientset, error) {
	if isInCluster() {
		return clientset.InCluster()
	}
	return clientset.OutOfCluster()
}

func lookupIntEnvVar(name string, def int) int {
	s, ok := os.LookupEnv(name)
	if !ok {
		return def
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return def
	}
	return int(v)
}

func isInCluster() bool {
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	return len(host) != 0 && len(port) != 0
}
