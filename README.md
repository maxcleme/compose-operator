# Compose-operator

PoC of trying to handle compose-spec "as-is" inside Kubernetes using an [Operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

All boilerplate/utilities is generated using [operator-framework](https://operatorframework.io/) for convenience, but everything could be done without if necessary.

Most of the logic is located at reconcile loop in [compose_controller.go](./controllers/compose_controller.go)

## Usage

#### Install CRD

```
make deploy
```

*Note: You can check if it runs correctly by looking at the following namespace.*

```
k get pod -n compose-operator-system
NAME                                                   READY   STATUS    RESTARTS   AGE
compose-operator-controller-manager-594654d948-w6ls4   2/2     Running   0          9m33s
```

#### Create one Compose resource

*k8s.compose.yaml*
```yaml
apiVersion: docker.com/v1alpha1
kind: Compose
metadata:
  name: test-compose
spec:
  spec: |
    name: maxcleme
    services:
      api:
        image: nginx:latest
        deploy:
          replicas: 5
      unknown:
        image: nginx:latest
        deploy:
          replicas: 2
```

```shell
k apply -f k8s.compose.yaml
```

```shell
k get pods -l project=maxcleme
NAME                                        READY   STATUS    RESTARTS   AGE
compose-maxcleme-api-67556d789d-98s5q       1/1     Running   0          11m
compose-maxcleme-api-67556d789d-bcrdk       1/1     Running   0          11m
compose-maxcleme-api-67556d789d-gplpm       1/1     Running   0          11m
compose-maxcleme-api-67556d789d-gq7qm       1/1     Running   0          11m
compose-maxcleme-api-67556d789d-t7zd4       1/1     Running   0          11m
compose-maxcleme-unknown-5d75546f7d-dkvbn   1/1     Running   0          11m
compose-maxcleme-unknown-5d75546f7d-xs5rl   1/1     Running   0          11m
```

## Limitation

For PoC sake, it only handles adding/removing services and replicas numbers.