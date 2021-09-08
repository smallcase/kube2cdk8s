# kube2cdk8s

A hacky way to generate cdk8s APIObject from a k8s yaml.

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
$ printf 'apiVersion: v1
kind: ServiceAccount
metadata:
  name: my-service-account
  namespace: my-namespace' > temp.yaml
```

```
$ go build
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
