# Pod best-by

## Installing the Chart

Before you can install the chart you will need to add the `bestby` repo to [Helm](https://helm.sh/).

```shell
helm repo add bestby https://pet2cattle.github.io/pod-best-by/
helm repo update
```

After you've installed the repo you can install the chart:

```shell
helm install -n bestby --create-namespace bestby bestby/bestby
```
