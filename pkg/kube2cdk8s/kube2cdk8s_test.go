package kube2cdk8s

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestKube2CDK8SServiceAccount(t *testing.T) {

	// create file that has a service account
	serviceAccount := `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: my-service-account
  namespace: my-namespace
`
	serviceAccountFile, err := createTempFile([]byte(serviceAccount))
	if err != nil {
		log.Println(err.Error())
	}

	d, err := Kube2CDK8S(serviceAccountFile.Name())
	if err != nil {
		log.Println(err.Error())
	}

	err = cupaloy.Snapshot(d)
	if err != nil {
		t.Error(err.Error())
	}

	defer os.Remove(serviceAccountFile.Name())
}

func TestKube2CDK8SDeployment(t *testing.T) {

	deployment := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
  namespace: my-namespace
spec:
  selector:
    matchLabels:
      app: my-deployment
  replicas: 3
  template:
    metadata:
      labels:
        app: my-deployment
    spec:
      containers:
      - name: my-deployment
        image: my-image
        imagePullPolicy: Always
        ports:
        - containerPort: 8080
`
	deploymentFile, err := createTempFile([]byte(deployment))
	if err != nil {
		log.Println(err.Error())
	}

	d, err := Kube2CDK8S(deploymentFile.Name())
	if err != nil {
		log.Println(err.Error())
	}

	err = cupaloy.Snapshot(d)
	if err != nil {
		t.Error(err.Error())
	}

	defer os.Remove(deploymentFile.Name())
}

func createTempFile(text []byte) (*os.File, error) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "prefix-")
	if err != nil {
		return nil, err
	}

	// Remember to clean up the file afterwards

	fmt.Println("Created File: " + tmpFile.Name())

	// Example writing to the file
	if _, err = tmpFile.Write(text); err != nil {
		return nil, err
	}

	// Close the file
	if err := tmpFile.Close(); err != nil {
		return nil, err
	}

	return tmpFile, nil
}
