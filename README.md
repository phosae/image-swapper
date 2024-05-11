# Image Swapper

Swap container image registry matched by prefix:

```
registry2registry:
  'registry.k8s.io': 'k8s.m.daocloud.io'
  'ghcr.io/k8snetworkplumbingwg': 'docker.io/myrepo'
```

In this example:
- registry.k8s.io/pause:3.8 -> k8s.m.daocloud.io/pause:3.8
- ghcr.io/k8snetworkplumbingwg/multus:v0 -> docker.io/myrepo/multus:v0

For a quick start, simply run `make`.

You can edit [Helm Chart Values](./charts/values.yaml) for customization.
