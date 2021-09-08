package kube2cdk8s

import (
	"log"
	"os"
	"testing"

	"github.com/smallcase/kube2cdk8s/util"

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
	serviceAccountFile, err := util.CreateTempFile([]byte(serviceAccount))
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
	deploymentFile, err := util.CreateTempFile([]byte(deployment))
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

func TestKube2CDK8SMultipleDeployment(t *testing.T) {

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
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment-2
  namespace: my-namespace-2
spec:
  selector:
    matchLabels:
      app: my-deployment-2
  replicas: 4
  template:
    metadata:
      labels:
        app: my-deployment-2
    spec:
      containers:
      - name: my-deployment-2
        image: my-image-2
        imagePullPolicy: Always
        ports:
        - containerPort: 8080
---
`
	deploymentFile, err := util.CreateTempFile([]byte(deployment))
	if err != nil {
		log.Println(err.Error())
	}

	d, err := Kube2CDK8SMultiple(deploymentFile.Name())
	if err != nil {
		log.Println(err.Error())
	}

	err = cupaloy.Snapshot(d)
	if err != nil {
		t.Error(err.Error())
	}

	defer os.Remove(deploymentFile.Name())
}

func TestKube2CDK8SMultipleDeploymentTwo(t *testing.T) {

	deployment := `---
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
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment-2
  namespace: my-namespace-2
spec:
  selector:
    matchLabels:
      app: my-deployment-2
  replicas: 4
  template:
    metadata:
      labels:
        app: my-deployment-2
    spec:
      containers:
      - name: my-deployment-2
        image: my-image-2
        imagePullPolicy: Always
        ports:
        - containerPort: 8080
---
`
	deploymentFile, err := util.CreateTempFile([]byte(deployment))
	if err != nil {
		log.Println(err.Error())
	}

	d, err := Kube2CDK8SMultiple(deploymentFile.Name())
	if err != nil {
		log.Println(err.Error())
	}

	err = cupaloy.Snapshot(d)
	if err != nil {
		t.Error(err.Error())
	}

	defer os.Remove(deploymentFile.Name())
}

func TestKube2CDK8SMultipleDeploymentThree(t *testing.T) {

	deployment := `---
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
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment-2
  namespace: my-namespace-2
spec:
  selector:
    matchLabels:
      app: my-deployment-2
  replicas: 4
  template:
    metadata:
      labels:
        app: my-deployment-2
    spec:
      containers:
      - name: my-deployment-2
        image: my-image-2
        imagePullPolicy: Always
        ports:
        - containerPort: 8080
`
	deploymentFile, err := util.CreateTempFile([]byte(deployment))
	if err != nil {
		log.Println(err.Error())
	}

	d, err := Kube2CDK8SMultiple(deploymentFile.Name())
	if err != nil {
		log.Println(err.Error())
	}

	err = cupaloy.Snapshot(d)
	if err != nil {
		t.Error(err.Error())
	}

	defer os.Remove(deploymentFile.Name())
}

func TestKube2CDK8SRole(t *testing.T) {

	role := `apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  labels:
    app.kubernetes.io/component: application-controller
    app.kubernetes.io/name: argocd-application-controller
    app.kubernetes.io/part-of: argocd
  name: argocd-application-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: argocd-application-controller
subjects:
- kind: ServiceAccount
  name: argocd-application-controller
`
	roleFile, err := util.CreateTempFile([]byte(role))
	if err != nil {
		log.Println(err.Error())
	}

	d, err := Kube2CDK8S(roleFile.Name())
	if err != nil {
		log.Println(err.Error())
	}

	err = cupaloy.Snapshot(d)
	if err != nil {
		t.Error(err.Error())
	}

	defer os.Remove(roleFile.Name())
}
