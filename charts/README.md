# kube-admission-webhook app Helm Chart

Installation

```shell
helm install image-swapper -n kube-system .
```

Upgrade

```shell
helm upgrade image-swapper -n kube-system --debug .
```

Uninstallation

```shell
helm -n kube-system uninstall image-swapper
```