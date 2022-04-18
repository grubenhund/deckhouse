/*
Copyright 2022 Flant JSC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package hooks

import (
	"strconv"

	"github.com/flant/addon-operator/pkg/module_manager/go_hook"
	"github.com/flant/addon-operator/pkg/module_manager/go_hook/metrics"
	"github.com/flant/addon-operator/sdk"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type etcdNode struct {
	Memory int64
	// isDedicated - indicate that node has taint
	//   - effect: NoSchedule
	//    key: node-role.kubernetes.io/master
	// it means that on node can be scheduled only control-plane components
	IsDedicated bool
}

const etcdBackendBytesGroup = "etcd_quota_backend_should_decrease"

var _ = sdk.RegisterFunc(&go_hook.HookConfig{
	Queue: etcdMaintenanceQueue,
	Kubernetes: []go_hook.KubernetesConfig{
		{
			Name:       "master_nodes",
			ApiVersion: "v1",
			Kind:       "Node",
			LabelSelector: &v1.LabelSelector{
				MatchLabels: map[string]string{
					"node-role.kubernetes.io/master": "",
				},
			},
			FilterFunc: etcdQuotaFilterNode,
		},
		etcdMaintenanceConfig,
	},
}, etcdQuotaBackendBytesHandler)

func etcdQuotaFilterNode(unstructured *unstructured.Unstructured) (go_hook.FilterResult, error) {
	var node corev1.Node

	err := sdk.FromUnstructured(unstructured, &node)
	if err != nil {
		return nil, err
	}

	memory := node.Status.Capacity.Memory().Value()

	isDedicated := false
	for _, taint := range node.Spec.Taints {
		if taint.Key == "node-role.kubernetes.io/master" && taint.Effect == corev1.TaintEffectNoSchedule {
			isDedicated = true
			break
		}
	}

	return &etcdNode{
		Memory:      memory,
		IsDedicated: isDedicated,
	}, nil
}

func getCurrentEtcdQuotaBytes(input *go_hook.HookInput) (int64, string) {
	var currentQuotaBytes int64
	var nodeWithMaxQuota string
	for _, endpointRaw := range input.Snapshots[etcdEndpointsSnapshotName] {
		endpoint := endpointRaw.(*etcdInstance)
		quotaForInstance := endpoint.MaxDbSize
		if quotaForInstance > currentQuotaBytes {
			currentQuotaBytes = quotaForInstance
			nodeWithMaxQuota = endpoint.Node
		}
	}

	if currentQuotaBytes == 0 {
		currentQuotaBytes = defaultEtcdMaxSize
		nodeWithMaxQuota = "default"
	}

	return currentQuotaBytes, nodeWithMaxQuota
}

func getNodeWithMinimalMemory(snapshots []go_hook.FilterResult) *etcdNode {
	if len(snapshots) == 0 {
		return nil
	}

	node := snapshots[0].(*etcdNode)
	for i := 1; i < len(snapshots); i++ {
		n := snapshots[i].(*etcdNode)
		// for not dedicated nodes we will not set new quota
		if !n.IsDedicated {
			return n
		}

		if n.Memory < node.Memory {
			node = n
		}
	}

	return node
}

func calcNewQuotaForMemory(minimalMemoryNodeBytes int64) int64 {
	const (
		minimalNodeSizeForCalc = 16 * 1024 * 1024 * 1024 // 24 GB
		nodeSizeStepForAdd     = 8 * 1024 * 1024 * 1024  // every 8 GB memory
		quotaStep              = 1 * 1024 * 1024 * 1024  // add 1 GB etcd memory every nodeSizeStepForAdd
		maxQuota               = 8 * 1024 * 1024 * 1024
	)

	if minimalMemoryNodeBytes <= minimalNodeSizeForCalc {
		return defaultEtcdMaxSize
	}

	steps := (minimalMemoryNodeBytes - minimalNodeSizeForCalc) / nodeSizeStepForAdd

	newQuota := steps*quotaStep + defaultEtcdMaxSize

	if newQuota > maxQuota {
		newQuota = maxQuota
	}

	return newQuota
}

func calcEtcdQuotaBackendBytes(input *go_hook.HookInput) int64 {
	currentQuotaBytes, nodeWithMaxQuota := getCurrentEtcdQuotaBytes(input)

	input.LogEntry.Debugf("Current etcd quota: %d. Getting from %s", currentQuotaBytes, nodeWithMaxQuota)

	node := getNodeWithMinimalMemory(input.Snapshots["master_nodes"])

	newQuotaBytes := currentQuotaBytes

	if node.IsDedicated {
		newQuotaBytes = calcNewQuotaForMemory(node.Memory)
		if newQuotaBytes < currentQuotaBytes {
			newQuotaBytes = currentQuotaBytes

			input.LogEntry.Warnf("Cannot decrease quota backend bytes. Current %d; calculated: %d. Use current", currentQuotaBytes, newQuotaBytes)

			input.MetricsCollector.Set(
				"d8_etcd_quota_backend_should_decrease",
				1.0,
				map[string]string{},
				metrics.WithGroup(etcdBackendBytesGroup))
		}

		input.LogEntry.Debugf("New backend quota bytes calculated: %d", newQuotaBytes)
	} else {
		input.LogEntry.Debugf("Found not dedicated control-plane node. Skip calculate backend quota. Use current: %d", newQuotaBytes)
	}

	return newQuotaBytes
}

func etcdQuotaBackendBytesHandler(input *go_hook.HookInput) error {
	input.MetricsCollector.Expire(etcdBackendBytesGroup)

	var newQuotaBytes int64

	userQuotaBytes := input.Values.Get("controlPlaneManager.etcd.maxDbSize")
	if userQuotaBytes.Exists() {
		newQuotaBytes = userQuotaBytes.Int()
	} else {
		newQuotaBytes = calcEtcdQuotaBackendBytes(input)
	}

	// use string because helm render big number in scientific format
	input.Values.Set("controlPlaneManager.internal.etcdQuotaBackendBytes", strconv.FormatInt(newQuotaBytes, 10))

	input.MetricsCollector.Set(
		"d8_etcd_quota_backend_total",
		float64(newQuotaBytes),
		map[string]string{})

	return nil
}
