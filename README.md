# kubewait

## Usage
```yaml
apiVersion: v1
kind: Pod
metadata:
    name: kubewait
    labels:
        app: kubewait
spec:
    restartPolicy: Never
    containers:
    - name: kubewait
      image: kubewait:latest
      imagePullPolicy: Never
      env:
      - name: KUBEWAIT
        value: |-
          [
            {
              "type": "Pod",
              "labelSelector": "app=postgres",
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
---

```

