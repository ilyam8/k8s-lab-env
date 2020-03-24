package bouncer

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/encryptio/alias"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiwatch "k8s.io/apimachinery/pkg/watch"
	clientappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/util/retry"
)

type Bouncer struct {
	LabelSelector string
	FieldSelector string
	RetryTimeout  time.Duration
	BounceEvery   time.Duration
	Client        clientappsv1.DeploymentInterface
	DryRun        bool

	tasks map[string]*task
}

func (b *Bouncer) Bounce() {
	if b.tasks == nil {
		b.tasks = make(map[string]*task)
	}

	for {
		if err := b.bounceOnce(); err != nil {
			log.WithField("error", err).
				WithField("reason", apierrors.ReasonForError(err)).
				Errorf("failed to bounce, will retry in %s", b.RetryTimeout)
		}
		b.stopTasks()
		time.Sleep(b.RetryTimeout)
	}
}

func (b *Bouncer) bounceOnce() error {
	watch, err := b.Client.Watch(metav1.ListOptions{
		LabelSelector: b.LabelSelector,
		FieldSelector: b.FieldSelector,
		Watch:         true,
	})
	if err != nil {
		return err
	}
	defer watch.Stop()

	for event := range watch.ResultChan() {
		b.handle(event)
	}
	return errors.New("watch channel was closed")
}

func (b *Bouncer) stopTasks() {
	for name, t := range b.tasks {
		<-t.stop()
		delete(b.tasks, name)
	}
}

func (b *Bouncer) handle(event apiwatch.Event) {
	d, ok := event.Object.(*appsv1.Deployment)
	if !ok {
		log.WithField("type", fmt.Sprintf("%T", event.Object)).Warnf("couldn't handle event, unknown object")
		return
	}

	if _, ok := d.GetLabels()["skipBounce"]; ok {
		return
	}

	switch event.Type {
	case apiwatch.Added:
		log.WithField("name", d.Name).Info("add deployment")
		b.handleAdded(d)
	case apiwatch.Deleted:
		log.WithField("name", d.Name).Info("delete deployment")
		b.handleDeleted(d)
	}
}

func (b *Bouncer) handleAdded(d *appsv1.Deployment) {
	if _, ok := b.tasks[d.Name]; !ok {
		f := func() { b.bounce(d.Name) }
		b.tasks[d.Name] = newTask(f, b.BounceEvery)
	}
}

func (b *Bouncer) handleDeleted(d *appsv1.Deployment) {
	if t, ok := b.tasks[d.Name]; ok {
		t.stop()
		delete(b.tasks, d.Name)
	}
}

func (b *Bouncer) bounce(name string) {
	err := retryOnConflict(func() error {
		if !doBounce() {
			log.WithField("name", name).Info("skip bouncing")
			return nil
		}

		d, err := b.Client.Get(name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		old := *d.Spec.Replicas
		switch addSub(*d.Spec.Replicas) {
		case add:
			*d.Spec.Replicas++
		case sub:
			*d.Spec.Replicas--
		}

		log.WithField("name", name).Infof("change replicas: %d => %d", old, *d.Spec.Replicas)
		_, err = b.Client.Update(d)
		return err
	})
	if err != nil {
		log.WithField("name", name).WithField("error", err).Info("failed to bounce deployment")
	}
}

func retryOnConflict(f func() error) error {
	return retry.RetryOnConflict(retry.DefaultRetry, f)
}

const maxReplicas = 3
const add = 0
const sub = 1

var (
	doBounceProbability = mustProbability(3, 7)
	doBounce            = func() bool { return doBounceProbability() == 0 }

	addSubProbability = mustProbability(5, 5)
	addSub            = func(current int32) uint32 {
		switch {
		case current >= maxReplicas:
			return sub
		case current <= 1:
			return add
		default:
			return addSubProbability()
		}
	}
)

func mustProbability(a, b float64, rest ...float64) func() uint32 {
	weights := append([]float64{a, b}, rest...)
	al, err := alias.New(weights)
	if err != nil {
		panic(err)
	}

	r := rand.New(rand.NewSource(time.Now().Unix()))
	return func() uint32 { return al.Gen(r) }
}
