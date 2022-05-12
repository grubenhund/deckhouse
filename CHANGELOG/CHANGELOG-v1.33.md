# Changelog v1.33

## [MALFORMED]


 - #1374 unknown section "global"
 - #1506 unknown section "monitoring"

## Features


 - **[candi]** Added Ubuntu 22.04 LTS support [#1505](https://github.com/deckhouse/deckhouse/pull/1505)
 - **[candi]** Bump containerd to v1.5.11. [#1386](https://github.com/deckhouse/deckhouse/pull/1386)
 - **[candi]** Improve candi bundle detection to detect centos-based distros [#1173](https://github.com/deckhouse/deckhouse/pull/1173)
 - **[cert-manager]** Add Cloudflare's APIToken support for ClusterIssuer [#1528](https://github.com/deckhouse/deckhouse/pull/1528)
 - **[cloud-provider-aws]** Added the ability to configure peering connections to the without-nat and standard layouts. [#514](https://github.com/deckhouse/deckhouse/pull/514)
 - **[cloud-provider-azure]** Enable accelerated networking for new machine-controller-manager instances. [#1266](https://github.com/deckhouse/deckhouse/pull/1266)
 - **[cloud-provider-yandex]** Changed default platform to standard-v3 for new instances created by machine-controller-manager. [#1361](https://github.com/deckhouse/deckhouse/pull/1361)
 - **[cni-flannel]** Bump flannel to 0.15.1. [#1173](https://github.com/deckhouse/deckhouse/pull/1173)
 - **[dhctl]** For new Deckhouse installations images for control-plane (image for pause container, for example) will be used from the Deckhouse registry. [#1517](https://github.com/deckhouse/deckhouse/pull/1517)
 - **[extended-monitoring]** List objects from the kube-apiserver cache, avoid hitting etcd on each list. It should decrease control plane resource consumption. [#1535](https://github.com/deckhouse/deckhouse/pull/1535)
 - **[helm]** Added deprecated APIs alerts for k8s 1.22 and 1.25 [#1461](https://github.com/deckhouse/deckhouse/pull/1461)
 - **[log-shipper]** Add label filter support for log-shipper. Users will be able to filter log messages based on their metadata labels. [#1424](https://github.com/deckhouse/deckhouse/pull/1424)
 - **[openvpn]** Added support for UDP protocol. [#1432](https://github.com/deckhouse/deckhouse/pull/1432)
 - **[prometheus]** Create table with enabled Deckhouse web interfaces on the Grafana home page [#1415](https://github.com/deckhouse/deckhouse/pull/1415)

## Fixes


 - **[candi]** Migrate to cgroupfs on containerd installations. [#1386](https://github.com/deckhouse/deckhouse/pull/1386)
 - **[helm]** Avoid hook failure on errors [#1523](https://github.com/deckhouse/deckhouse/pull/1523)
 - **[kube-dns]** Updated CoreDNS to v1.9.1 [#1537](https://github.com/deckhouse/deckhouse/pull/1537)
 - **[log-shipper]** Migrate deprecated elasticsearch fields [#1453](https://github.com/deckhouse/deckhouse/pull/1453)
 - **[log-shipper]** Send reloading signal to all vector processes in a container on config change. [#1430](https://github.com/deckhouse/deckhouse/pull/1430)
 - **[monitoring-kubernetes]** Fix kubelet alerts [#1471](https://github.com/deckhouse/deckhouse/pull/1471)
 - **[node-local-dns]** Updated CoreDNS to v1.9.1 [#1537](https://github.com/deckhouse/deckhouse/pull/1537)
 - **[prometheus]** Removed the old prometheus_storage_class_change shell hook which has already been replaced by Go hooks. [#1396](https://github.com/deckhouse/deckhouse/pull/1396)
 - **[upmeter]** UI shows only present data [#1405](https://github.com/deckhouse/deckhouse/pull/1405)
 - **[upmeter]** Use finite timeout in agent insecure HTTP client [#1334](https://github.com/deckhouse/deckhouse/pull/1334)
 - **[upmeter]** Fixed slow data loading in [#1257](https://github.com/deckhouse/deckhouse/pull/1257)

## Chore


 - **[dashboard]** Dashboard upgrade from 2.2.0 to 2.5.1 [#1383](https://github.com/deckhouse/deckhouse/pull/1383)
 - **[docs]** Suggest gp3 for bastion instance in AWS-based 'Getting started' [#1495](https://github.com/deckhouse/deckhouse/pull/1495)
 - **[flant-integration]** Removed unused "flantIntegration.team" field from values schema [#1514](https://github.com/deckhouse/deckhouse/pull/1514)
 - **[ingress-nginx]** Bump GoGo dependency for the protobuf-exporter to prevent improper input. [#1519](https://github.com/deckhouse/deckhouse/pull/1519)

