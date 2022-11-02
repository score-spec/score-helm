# 02 - Environment Variables

It is common to pass a dynamic environment-specific configuration to the container during the deployment.

In score specification a special `environment` resource type is used to support such cases:

```yaml
apiVersion: score.dev/v1b1

metadata:
  name: hello-world

containers:
  hello:
    image: busybox
    command: ["/bin/sh"]
    args: ["-c", "while true; do echo Hello $${FRIEND}!; sleep 5; done"]
    variables:
      FRIEND: ${resources.env.NAME}

resources:
  env:
    type: environment
    properties:
      NAME:
        type: string
        default: World
```

If there is a need to set the `NAME` variable value for the environment the workload would be deployed into, an `env.yaml` file can be used:

```yaml
env:
  NAME: John
```

Now the source `score.yaml` file and the `env.yaml` file should be combined and converted into Helm values file with `score-helm` CLI tool:

```console
$ score-helm run -f ./score.yaml --values ./env.yaml -o ./values.yaml
```

Output `values.yaml` file would contain a workload configuration for the chart:

```yaml
containers:
  hello:
    args:
      - -c
      - while true; do echo Hello $${FRIEND}!; sleep 5; done
    command:
      - /bin/sh
    env:
      - name: FRIEND
        value: John
    image:
      name: busybox
```

A generic chart template from `/examples/chart` can be used to prepare a Helm chart for this example:

```console
$ helm create -p ../examples/chart hello
```

Deploying the Helm chart:

```console
$ helm install --values ./values.yaml hello ./hello

NAME: hello
LAST DEPLOYED: Wed Nov 2 14:40:11 2022
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None
```

Resulting Kubernetes deployment object:

```yaml
# Source: hello/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello
  labels:
    helm.sh/chart: hello-0.1.0
    app.kubernetes.io/name: 
    app.kubernetes.io/instance: hello
    app.kubernetes.io/version: "0.1.0"
    app.kubernetes.io/managed-by: Helm
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: 
      app.kubernetes.io/instance: hello
  template:
    metadata:
      labels:
        app.kubernetes.io/name: 
        app.kubernetes.io/instance: hello
    spec:
      containers:
        - name: hello
          image: "busybox"
          command:
            - /bin/sh
          args:
            - -c
            - while true; do echo Hello $${FRIEND}!; sleep 5; done
          env:
            - name: FRIEND
              value: John
```
