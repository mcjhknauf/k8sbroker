## Broker
A simple HTTP server to create k8s Jobs from a CronJob

### Requirements
- A local k8s cluster, like minikube (`minikube start --kubernetes-version=v1.22.4`)

### Create the Cronjob
```
kubectl apply -f cronjob.yml
```

### Run the HTTP Server
```
go run main.go
```

### Create a Job
```
curl -v -X POST -d 'name=test' http://localhost:8000/job
```
