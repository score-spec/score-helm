# 01 - Hello World!

In this basic example there is a simple workload based on `busybox` Docker image described in a `score.yaml` file:

```yaml
apiVersion: score.dev/v1b1

metadata:
  name: hello-world

containers:
  hello:
    image: busybox
    command: ["/bin/sh"]
    args: ["-c", "while true; do echo Hello World!; sleep 5; done"]
```

A generic chart template from `/examples/chart` can be used to prepare a Helm chart for this workload:

```console
$ helm create -p ../examples/chart hello
```

Now the source `score.yaml` file should be converted into Helm values file with `score-helm` CLI tool:

```console
$ score-helm run -f ./score.yaml -o ./values.yaml
```

Output `values.yaml` file would contain a workload configuration for the chart:

```yaml
containers:
  hello:
    args:
      - -c
      - while true; do echo Hello World!; sleep 5; done
    command:
      - /bin/sh
    image:
      name: busybox
```

Deploying the Helm chart:

```console
$ helm install --values ./values.yaml hello ./hello

NAME: hello
LAST DEPLOYED: Mon Oct 31 16:06:34 2022
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
            - while true; do echo Hello World!; sleep 5; done
```
