package main

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/op/go-logging"

	"k8s.io/client-go/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	lifetime       string = "pod.kubernetes.io/lifetime"
	ignorelifetime string = "pod.kubernetes.io/ignore-lifetime"
)

var log = logging.MustGetLogger("shelf-stocker")

var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.5s}%{color:reset} %{message}`,
)

func main() {
	backend := logging.NewLogBackend(os.Stderr, "", 0)

	formatted_backed := logging.NewBackendFormatter(backend, format)

	leveled_backend := logging.AddModuleLevel(formatted_backed)
	if isDebug() {
		leveled_backend.SetLevel(logging.DEBUG, "")
	} else {
		leveled_backend.SetLevel(logging.INFO, "")
	}

	logging.SetBackend(leveled_backend)

	ctx := context.Background()
	config := ctrl.GetConfigOrDie()
	clientset := kubernetes.NewForConfigOrDie(config)

	namespace_list, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	log.Info("initializing shelf-stocker")

	if useAnnotations() {
		log.Info("using annotations")
	} else {
		log.Info("using labels")
	}

	for {
		for _, namespace := range namespace_list.Items {
			killed_pods := 0

			// exclude mode
			if useAnnotations() && namespace.Annotations[ignorelifetime] == "true" {
				log.Debugf("found ignorelifetime: skipping namespace %+v\n", namespace.Name)
				continue
			} else if !useAnnotations() && namespace.Labels[ignorelifetime] == "true" {
				log.Debugf("found ignorelifetime: skipping namespace %+v\n", namespace.Name)
				continue
			}

			// process Pods
			pod_list, err := clientset.CoreV1().Pods(namespace.Name).List(ctx, metav1.ListOptions{})
			if err != nil {
				continue
			}

			for _, pod := range pod_list.Items {

				log.Debugf("considering: namespace %s Pod %s", namespace.Name, pod.Name)

				var val string
				var ok bool

				if useAnnotations() {
					val, ok = pod.Annotations[lifetime]
				} else {
					val, ok = pod.Labels[lifetime]
				}

				if ok {
					lifetime := time.Second
					minutes, err := strconv.Atoi(val)

					if err != nil {
						lifetime, _ = time.ParseDuration(val)
					} else {
						lifetime = time.Duration(minutes) * time.Minute
					}

					if lifetime == 0 {
						log.Infof("skipping Pod: namespace %s pod %s : provided value %s is incorrect\n", namespace.Name, pod.Name, val)
					} else if pod.Status.Phase == "Running" {

						log.Debugf("namespace", namespace.Name, "pod", pod.Name, "is running")
						log.Debugf("start time", pod.Status.StartTime)
						log.Debugf("kill time", pod.Status.StartTime.Add(lifetime))

						if pod.Status.StartTime.Add(lifetime).Before(time.Now()) {
							if maxKilledPods() > 0 && killed_pods < maxKilledPods() {
								// kill pod
								err := clientset.CoreV1().Pods(namespace.Name).Delete(ctx, pod.Name, metav1.DeleteOptions{})
								if err != nil {
									log.Errorf("ERROR killing Pod: namespace %s pod %s : %s\n", namespace.Name, pod.Name, err.Error())
								} else {
									log.Infof("Pod KILLED: namespace %s pod %s\n", namespace.Name, pod.Name)
									killed_pods++
								}
							} else {
								log.Warningf("skipping Pods for namespace %s: max killed pods reached %d\n", namespace.Name, maxKilledPods())
								break
							}
						}
					}
				}
			}
		}

		if h := os.Getenv("RUN_ONCE"); h != "" {
			break
		}

		log.Infof("sleeping for %d seconds\n", int(sleepDuration().Seconds()))
		time.Sleep(sleepDuration())

	}
}

func sleepDuration() time.Duration {
	if h := os.Getenv("INTERVAL_IN_SEC"); h != "" {
		s, _ := strconv.Atoi(h)
		return time.Duration(s) * time.Second
	}
	return 60 * time.Second
}

func maxKilledPods() int {
	if h := os.Getenv("MAX_KILLED_PODS_NS"); h != "" {
		s, _ := strconv.Atoi(h)
		return s
	}
	return 5
}

func isDebug() bool {
	if h := os.Getenv("DEBUG"); h != "" {
		s, _ := strconv.Atoi(h)
		return s > 0
	}
	return false
}

func useAnnotations() bool {
	if h := os.Getenv("ANNOTATIONS"); h != "" {
		s, _ := strconv.Atoi(h)
		return s > 0
	}
	return false
}
