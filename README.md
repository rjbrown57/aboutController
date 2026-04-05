# aboutcontroller

aboutController is a simple controller to [about-api](https://github.com/kubernetes-sigs/about-api) clusterProperites based on annotations to k8s workload objects.

To utilize this controller add any number of annotations to your workload with the prefix `aboutcontroller.io`.

The name of the generated properties will be what comes after the /, and the value will be whatever the annotation value is. 

Examples
* `aboutcontroller.io/myapp: "v0.0.0"` will result in a clusterProperty being created with the content

```
apiVersion: about.k8s.io/v1alpha1
kind: ClusterProperty
metadata:
  name: myapp
spec:
  value: v0.0.0
```

On removal of the workload that generated the property all properties will be removed.

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

