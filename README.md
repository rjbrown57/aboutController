# aboutcontroller

`aboutcontroller` is a Kubernetes controller that creates [about-api](https://github.com/kubernetes-sigs/about-api) `ClusterProperty` resources from annotations on workload objects.

## How It Works

Add one or more annotations to a supported workload with the prefix `aboutcontroller.io/`.

The controller currently watches:

- `Deployment`
- `StatefulSet`
- `DaemonSet`

For each matching annotation:

- the property name is the part after `aboutcontroller.io/`
- the property value is the annotation value

For example, this annotation:

```yaml
aboutcontroller.io/myapp: "v0.0.0"
```

results in a `ClusterProperty` like:

```yaml
apiVersion: about.k8s.io/v1alpha1
kind: ClusterProperty
metadata:
  name: myapp
spec:
  value: v0.0.0
```

## Reconciliation Behavior

The controller will:

- create properties for matching annotations
- update properties when annotation values change
- delete properties when the annotation is removed from the workload
- delete managed properties when the workload itself is deleted

## Examples

Example manifests are available in [`examples/`](/Users/lookfar/repos/github.com/rjbrown57/aboutController/examples):

- [`deployment.yaml`](/Users/lookfar/repos/github.com/rjbrown57/aboutController/examples/deployment.yaml)
- [`statefulset.yaml`](/Users/lookfar/repos/github.com/rjbrown57/aboutController/examples/statefulset.yaml)
- [`daemonset.yaml`](/Users/lookfar/repos/github.com/rjbrown57/aboutController/examples/daemonset.yaml)

## License

Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
