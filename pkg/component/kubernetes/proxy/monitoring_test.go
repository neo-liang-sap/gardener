// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package proxy_test

import (
	. "github.com/onsi/ginkgo/v2"

	"github.com/gardener/gardener/pkg/component"
	. "github.com/gardener/gardener/pkg/component/kubernetes/proxy"
	"github.com/gardener/gardener/pkg/component/test"
)

var _ = Describe("Monitoring", func() {
	var component component.MonitoringComponent

	BeforeEach(func() {
		component = New(nil, "", Values{})
	})

	It("should successfully test the scrape config", func() {
		test.ScrapeConfigs(component, expectedScrapeConfig)
	})

	It("should successfully test the alerting rules", func() {
		test.AlertingRulesWithPromtool(
			component,
			map[string]string{"kube-proxy.rules.yaml": expectedAlertingRule},
			"",
		)
	})
})

const (
	expectedScrapeConfig = `job_name: kube-proxy
honor_labels: false
scheme: https
tls_config:
  ca_file: /etc/prometheus/seed/ca.crt
authorization:
  type: Bearer
  credentials_file: /var/run/secrets/gardener.cloud/shoot/token/token
follow_redirects: false
kubernetes_sd_configs:
- role: endpoints
  api_server: https://kube-apiserver:443
  namespaces:
    names: [ kube-system ]
  tls_config:
    ca_file: /etc/prometheus/seed/ca.crt
  authorization:
    type: Bearer
    credentials_file: /var/run/secrets/gardener.cloud/shoot/token/token
relabel_configs:
- source_labels:
  - __meta_kubernetes_endpoints_name
  - __meta_kubernetes_endpoint_port_name
  action: keep
  regex: kube-proxy;metrics
- source_labels: [ __meta_kubernetes_pod_name ]
  target_label: pod
- source_labels: [ __meta_kubernetes_pod_node_name ]
  target_label: node
- target_label: __address__
  replacement: kube-apiserver:443
- source_labels: [__meta_kubernetes_pod_name, __meta_kubernetes_pod_container_port_number]
  regex: (.+);(.+)
  target_label: __metrics_path__
  replacement: /api/v1/namespaces/kube-system/pods/${1}:${2}/proxy/metrics
metric_relabel_configs:
- source_labels: [ __name__ ]
  action: keep
  regex: ^(kubeproxy_network_programming_duration_seconds_bucket|kubeproxy_network_programming_duration_seconds_count|kubeproxy_network_programming_duration_seconds_sum|kubeproxy_sync_proxy_rules_duration_seconds_bucket|kubeproxy_sync_proxy_rules_duration_seconds_count|kubeproxy_sync_proxy_rules_duration_seconds_sum)$
`

	expectedAlertingRule = `groups:
- name: kube-proxy.rules
  rules:
  - record: kubeproxy_network_latency:quantile
    expr: histogram_quantile(0.99, sum(rate(kubeproxy_network_programming_duration_seconds_bucket[10m])) by (le))
    labels:
      quantile: "0.99"
  - record: kubeproxy_network_latency:quantile
    expr: histogram_quantile(0.9, sum(rate(kubeproxy_network_programming_duration_seconds_bucket[10m])) by (le))
    labels:
      quantile: "0.9"
  - record: kubeproxy_network_latency:quantile
    expr: histogram_quantile(0.5, sum(rate(kubeproxy_network_programming_duration_seconds_bucket[10m])) by (le))
    labels:
      quantile: "0.5"
  - record: kubeproxy_sync_proxy:quantile
    expr: histogram_quantile(0.99, sum(rate(kubeproxy_sync_proxy_rules_duration_seconds_bucket[10m])) by (le))
    labels:
      quantile: "0.99"
  - record: kubeproxy_sync_proxy:quantile
    expr: histogram_quantile(0.9, sum(rate(kubeproxy_sync_proxy_rules_duration_seconds_bucket[10m])) by (le))
    labels:
      quantile: "0.9"
  - record: kubeproxy_sync_proxy:quantile
    expr: histogram_quantile(0.5, sum(rate(kubeproxy_sync_proxy_rules_duration_seconds_bucket[10m])) by (le))
    labels:
      quantile: "0.5"
`
)
