# kube2cdk8s

Converts your k8s YAML to a cdk8s Api Object.

Uses Pulumi's kube2pulumi as a base.

## Dependencies

```
1. pulumi cli
2. pulumi kubernetes provider
```

```
$ curl -fsSL https://get.pulumi.com | sh
$ pulumi plugin install resource kubernetes v2.4.2
```

## Usage

```
$ go test
$ go build
```
```
$ printf 'apiVersion: v1
kind: ServiceAccount
metadata:
  name: my-service-account
  namespace: my-namespace' > temp.yaml
```
```
$ ./kube2cdk8s typescript -f temp.yaml
new cdk8s.ApiObject("", this, {
    apiVersion: "v1",
    kind: "ServiceAccount",
    metadata: {
        name: "my-service-account",
        namespace: "my-namespace",
    },
});
```
```
printf '---
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
        - containerPort: 8080' > temp.yaml
```
```
$ ./kube2cdk8s typescript -m true -f temp.yaml
new cdk8s.ApiObject("", this, {
    apiVersion: "apps/v1",
    kind: "Deployment",
    metadata: {
        name: "my-deployment",
        namespace: "my-namespace",
    },
    spec: {
        selector: {
            matchLabels: {
                app: "my-deployment",
            },
        },
        replicas: 3,
        template: {
            metadata: {
                labels: {
                    app: "my-deployment",
                },
            },
            spec: {
                containers: [{
                    name: "my-deployment",
                    image: "my-image",
                    imagePullPolicy: "Always",
                    ports: [{
                        containerPort: 8080,
                    }],
                }],
            },
        },
    },
});

new cdk8s.ApiObject("", this, {
    apiVersion: "apps/v1",
    kind: "Deployment",
    metadata: {
        name: "my-deployment-2",
        namespace: "my-namespace-2",
    },
    spec: {
        selector: {
            matchLabels: {
                app: "my-deployment-2",
            },
        },
        replicas: 4,
        template: {
            metadata: {
                labels: {
                    app: "my-deployment-2",
                },
            },
            spec: {
                containers: [{
                    name: "my-deployment-2",
                    image: "my-image-2",
                    imagePullPolicy: "Always",
                    ports: [{
                        containerPort: 8080,
                    }],
                }],
            },
        },
    },
});
```
