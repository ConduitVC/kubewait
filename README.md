# kubewait

Kubewait can be used as an `initContainer` to allow a Pod/Job to wait on another kubernetes (or external, maybe) resource.
Kubewait takes a list of `StateDescription` objects and waits until the cluster state matches that description.
`StateDescription` consists of the following fields:
1. `type: String`: The type of resource to be monitored. It can be `Pod` or `Job`.
2. `labelSelector: String`: A kubernetes `LabelSelector` for the required resource.
3. `requiredStates: [ String ]`: Matches if the resource is in one of these states.
4. `namespace`: Namespace of the resource.

| `type` | allowed values in `requiredStates` |
|---|---|
| Pod | `Ready`, `Succeeded`, `Failed`|
| Job | `Running`, `Complete`, `Failed` |

## RBAC
`kubewait` requires permissions to watch the states of pods/jobs. To grant permissions for kubewait in a single namespace:
```yaml
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  namespace: default
  name: kubewait
rules:
- apiGroups: ["", "batch"] # "" indicates the core API group
  resources: ["pods", "jobs"]
  verbs: ["get", "watch", "list"]
---
# Every namespace has a service account called default

kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: kubewait
  namespace: example-ns
subjects:
- kind: ServiceAccount
  name: default
  namespace: example-ns
roleRef:
  kind: Role
  name: kubewait
  namespace: example-ns
```

## Example
Consider an app which depends on postgres (which needs to be seeded) and redis.
```yaml
---
# Create a database deployment

apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres 
  labels:
    app: postgres
spec:
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
      ...
---
# Create a redis deployment

apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
spec:
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
      ...
---
# Create the seeder job

apiVersion: batch/v1
kind: Job
metadata:
  name: seeder
  labels:
    app: seeder
spec:
  template:
    spec:
      initContainers:
      # Wait for the database to be ready
      - name: kubewait
        image: ckousik/kubewait:latest
        env:
        - name: KUBEWAIT
          value: |-
            [
              {
                "type": "Pod",,
                "labelSelector": "app=postgres",
                "requiredStates": [ "Ready" ],
                "namespace": "default"
              },
            ]
      containers:
      - name: seeder
      ....
---
apiVersion: v1
kind: Pod
metadata:
  name: myapp
  labels:
    app: myapp
spec:
  restartPolicy: Never
  initContainers:
  - name: kubewait
    image: ckousik/kubewait:latest
    env:
    - name: KUBEWAIT
      value: |-
        [
          {
            "type": "Pod",
            "labelSelector": "app=redis",
            "requiredStates": [ "Ready" ],
            "namespace": "default"
          },
          {
            "type": "Job",
            "labelSelector": "app=seeder",
            "requiredStates": ["Completed"],
            "namespace": "default"
          }
        ]
    - name: ENV
      value: "DEBUG"
  containers:
  - name: myapp
    ...
---

```

