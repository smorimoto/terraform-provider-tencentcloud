---
subcategory: "Managed Service for Prometheus(TMP)"
layout: "tencentcloud"
page_title: "TencentCloud: tencentcloud_monitor_tmp_exporter_integration"
sidebar_current: "docs-tencentcloud-resource-monitor_tmp_exporter_integration"
description: |-
  Provides a resource to create a monitor tmpExporterIntegration
---

# tencentcloud_monitor_tmp_exporter_integration

Provides a resource to create a monitor tmpExporterIntegration

## Example Usage

```hcl
resource "tencentcloud_monitor_tmp_exporter_integration" "tmpExporterIntegration" {
  instance_id = "prom-dko9d0nu"
  kind        = "blackbox-exporter"
  content     = "{\"name\":\"test\",\"kind\":\"blackbox-exporter\",\"spec\":{\"instanceSpec\":{\"module\":\"http_get\",\"urls\":[\"xx\"]}}}"
  kube_type   = 1
  cluster_id  = "cls-bmuaukfu"
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required, String) Cluster ID.
* `content` - (Required, String) Integration config.
* `instance_id` - (Required, String) Instance id.
* `kind` - (Required, String) Type.
* `kube_type` - (Required, Int) Integration config.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.



