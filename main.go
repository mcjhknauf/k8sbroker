package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	logger := log.New(os.Stdout, "broker: ", log.LstdFlags|log.Lmsgprefix)

	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	config, err := clientcmd.BuildConfigFromFlags("", path.Join(home, ".kube/config"))
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	batchClient := clientset.BatchV1()

	cronJob, err := batchClient.CronJobs("default").Get(context.TODO(), "hello", metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	annotations := make(map[string]string)
	annotations["cronjob.kubernetes.io/instantiate"] = "manual"
	for k, v := range cronJob.Spec.JobTemplate.Annotations {
		annotations[k] = v
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/job", func(w http.ResponseWriter, r *http.Request) {
		job := &batchv1.Job{
			// this is ok because we know exactly how we want to be serialized
			TypeMeta: metav1.TypeMeta{APIVersion: batchv1.SchemeGroupVersion.String(), Kind: "Job"},
			ObjectMeta: metav1.ObjectMeta{
				Name:        r.FormValue("name"),
				Annotations: annotations,
				Labels:      cronJob.Spec.JobTemplate.Labels,
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: batchv1.SchemeGroupVersion.String(),
						Kind:       "CronJob",
						Name:       cronJob.GetName(),
						UID:        cronJob.GetUID(),
					},
				},
				Namespace: "default",
			},
			Spec: cronJob.Spec.JobTemplate.Spec,
		}

		_, err = batchClient.Jobs("default").Create(context.TODO(), job, metav1.CreateOptions{})
		if err != nil {
			logger.Println(err)
		}
	})

	server := http.Server{
		Addr:         ":8000",
		Handler:      mux,
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 10,
		IdleTimeout:  time.Second * 120,
		ErrorLog:     logger,
	}

	logger.Fatal(server.ListenAndServe())
}
