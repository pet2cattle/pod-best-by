# Pod best-by

## Installing the Chart

Before you can install the chart you will need to add the `stevehipwell` repo to [Helm](https://helm.sh/).

```shell
helm repo add pet2cattle https://jordiprats.github.io/pet2cattle-pod-best-by/
```

After you've installed the repo you can install the chart:

```shell
helm upgrade --install --namespace default bestby pet2cattle/bestby
```
