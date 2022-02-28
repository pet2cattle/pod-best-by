package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	// "k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	lifetimeLabel       string = "pod.kubernetes.io/lifetime"
	ignorelifetimeLabel string = "pod.kubernetes.io/ignore-lifetime"
)

func main() {
	ctx := context.Background()
	config := ctrl.GetConfigOrDie()
	clientset := kubernetes.NewForConfigOrDie(config)

	namespace_list, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for {
		for _, namespace := range namespace_list.Items {
			killed_pods := 0

			if val, ok := namespace.Labels[ignorelifetimeLabel]; ok {
				if val == "true" {
					fmt.Printf("%+v\n", namespace.Name)
					continue
				}
			}

			// process Pods
			pod_list, err := clientset.CoreV1().Pods(namespace.Name).List(ctx, metav1.ListOptions{})
			if err != nil {
				continue
			}

			for _, pod := range pod_list.Items {
				// fmt.Println("considering: namespace", namespace.Name, "pod", pod.Name)
				if val, ok := pod.Labels[lifetimeLabel]; ok {
					lifetime := time.Second
					minutes, err := strconv.Atoi(val)

					if err != nil {
						lifetime, _ = time.ParseDuration(val)
					} else {
						lifetime = time.Duration(minutes) * time.Minute
					}

					if lifetime == 0 {
						fmt.Printf("namespace %s pod %s : provided value %s is incorrect\n", namespace.Name, pod.Name, val)
					} else if pod.Status.Phase == "Running" {
						fmt.Println("namespace", namespace.Name, "pod", pod.Name, "is running")
						// show start time
						fmt.Println("start time", pod.Status.StartTime)
						fmt.Println("kill time", pod.Status.StartTime.Add(lifetime))
						if pod.Status.StartTime.Add(lifetime).Before(time.Now()) {
							if maxKilledPods() > 0 && killed_pods < maxKilledPods() {
								// kill pod
								err := clientset.CoreV1().Pods(namespace.Name).Delete(ctx, pod.Name, metav1.DeleteOptions{})
								if err != nil {
									fmt.Printf("namespace %s pod %s : %s\n", namespace.Name, pod.Name, err.Error())
								} else {
									fmt.Printf("namespace %s KILLED pod %s\n", namespace.Name, pod.Name)
									killed_pods++
								}
							} else {
								fmt.Printf("namespace %s: max killed pods reached %d\n", namespace.Name, maxKilledPods())
							}
						}
					}
				}
			}
		}

		if h := os.Getenv("RUN_ONCE"); h != "" {
			break
		}

		fmt.Printf("Now sleeping for %d seconds", int(sleepDuration().Seconds()))
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
