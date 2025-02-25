/*
Provides a resource to create a cfs snapshot

Example Usage

```hcl
resource "tencentcloud_cfs_snapshot" "snapshot" {
  file_system_id = "cfs-iobiaxtj"
  snapshot_name = "test"
  tags = {
    "createdBy" = "terraform"
  }
}
```

Import

cfs snapshot can be imported using the id, e.g.

```
terraform import tencentcloud_cfs_snapshot.snapshot snapshot_id
```
*/
package tencentcloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	cfs "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cfs/v20190719"
	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/internal/helper"
)

func resourceTencentCloudCfsSnapshot() *schema.Resource {
	return &schema.Resource{
		Create: resourceTencentCloudCfsSnapshotCreate,
		Read:   resourceTencentCloudCfsSnapshotRead,
		Update: resourceTencentCloudCfsSnapshotUpdate,
		Delete: resourceTencentCloudCfsSnapshotDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"file_system_id": {
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
				Description: "Id of file system.",
			},

			"snapshot_name": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "Name of snapshot.",
			},

			"tags": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Tag description list.",
			},
		},
	}
}

func resourceTencentCloudCfsSnapshotCreate(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_cfs_snapshot.create")()
	defer inconsistentCheck(d, meta)()

	logId := getLogId(contextNil)

	var (
		request    = cfs.NewCreateCfsSnapshotRequest()
		response   = cfs.NewCreateCfsSnapshotResponse()
		snapshotId string
	)
	if v, ok := d.GetOk("file_system_id"); ok {
		request.FileSystemId = helper.String(v.(string))
	}

	if v, ok := d.GetOk("snapshot_name"); ok {
		request.SnapshotName = helper.String(v.(string))
	}

	err := resource.Retry(writeRetryTimeout, func() *resource.RetryError {
		result, e := meta.(*TencentCloudClient).apiV3Conn.UseCfsClient().CreateCfsSnapshot(request)
		if e != nil {
			return retryError(e)
		} else {
			log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), result.ToJsonString())
		}
		response = result
		return nil
	})
	if err != nil {
		log.Printf("[CRITAL]%s create cfs snapshot failed, reason:%+v", logId, err)
		return err
	}

	snapshotId = *response.Response.SnapshotId
	d.SetId(snapshotId)

	service := CfsService{client: meta.(*TencentCloudClient).apiV3Conn}

	conf := BuildStateChangeConf([]string{}, []string{"available"}, 2*readRetryTimeout, time.Second, service.CfsSnapshotStateRefreshFunc(d.Id(), []string{}))

	if _, e := conf.WaitForState(); e != nil {
		return e
	}

	ctx := context.WithValue(context.TODO(), logIdKey, logId)
	if tags := helper.GetTags(d, "tags"); len(tags) > 0 {
		tagService := TagService{client: meta.(*TencentCloudClient).apiV3Conn}
		region := meta.(*TencentCloudClient).apiV3Conn.Region
		resourceName := fmt.Sprintf("qcs::cfs:%s:uin/:snap/%s", region, d.Id())
		if err := tagService.ModifyTags(ctx, resourceName, tags, nil); err != nil {
			return err
		}
	}

	return resourceTencentCloudCfsSnapshotRead(d, meta)
}

func resourceTencentCloudCfsSnapshotRead(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_cfs_snapshot.read")()
	defer inconsistentCheck(d, meta)()

	logId := getLogId(contextNil)

	ctx := context.WithValue(context.TODO(), logIdKey, logId)

	service := CfsService{client: meta.(*TencentCloudClient).apiV3Conn}

	snapshotId := d.Id()

	snapshot, err := service.DescribeCfsSnapshotById(ctx, snapshotId)
	if err != nil {
		return err
	}

	if snapshot == nil {
		d.SetId("")
		log.Printf("[WARN]%s resource `CfsSnapshot` [%s] not found, please check if it has been deleted.\n", logId, d.Id())
		return nil
	}

	if snapshot.FileSystemId != nil {
		_ = d.Set("file_system_id", snapshot.FileSystemId)
	}

	if snapshot.SnapshotName != nil {
		_ = d.Set("snapshot_name", snapshot.SnapshotName)
	}

	tcClient := meta.(*TencentCloudClient).apiV3Conn
	tagService := &TagService{client: tcClient}
	tags, err := tagService.DescribeResourceTags(ctx, "cfs", "snap", tcClient.Region, d.Id())
	if err != nil {
		return err
	}
	_ = d.Set("tags", tags)

	return nil
}

func resourceTencentCloudCfsSnapshotUpdate(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_cfs_snapshot.update")()
	defer inconsistentCheck(d, meta)()

	logId := getLogId(contextNil)
	needChange := false

	request := cfs.NewUpdateCfsSnapshotAttributeRequest()

	snapshotId := d.Id()

	request.SnapshotId = &snapshotId

	if d.HasChange("snapshot_name") {
		needChange = true
		if v, ok := d.GetOk("snapshot_name"); ok {
			request.SnapshotName = helper.String(v.(string))
		}
	}

	if needChange {
		err := resource.Retry(writeRetryTimeout, func() *resource.RetryError {
			result, e := meta.(*TencentCloudClient).apiV3Conn.UseCfsClient().UpdateCfsSnapshotAttribute(request)
			if e != nil {
				return retryError(e)
			} else {
				log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), result.ToJsonString())
			}
			return nil
		})
		if err != nil {
			log.Printf("[CRITAL]%s update cfs snapshot failed, reason:%+v", logId, err)
			return err
		}
	}

	if d.HasChange("tags") {
		ctx := context.WithValue(context.TODO(), logIdKey, logId)
		tcClient := meta.(*TencentCloudClient).apiV3Conn
		tagService := &TagService{client: tcClient}
		oldTags, newTags := d.GetChange("tags")
		replaceTags, deleteTags := diffTags(oldTags.(map[string]interface{}), newTags.(map[string]interface{}))
		resourceName := BuildTagResourceName("cfs", "snap", tcClient.Region, d.Id())
		if err := tagService.ModifyTags(ctx, resourceName, replaceTags, deleteTags); err != nil {
			return err
		}
	}

	return resourceTencentCloudCfsSnapshotRead(d, meta)
}

func resourceTencentCloudCfsSnapshotDelete(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_cfs_snapshot.delete")()
	defer inconsistentCheck(d, meta)()

	logId := getLogId(contextNil)
	ctx := context.WithValue(context.TODO(), logIdKey, logId)

	service := CfsService{client: meta.(*TencentCloudClient).apiV3Conn}
	snapshotId := d.Id()

	if err := service.DeleteCfsSnapshotById(ctx, snapshotId); err != nil {
		return err
	}

	err := resource.Retry(2*readRetryTimeout, func() *resource.RetryError {
		instance, errRet := service.DescribeCfsSnapshotById(ctx, snapshotId)
		if errRet != nil {
			return retryError(errRet, InternalError)
		}
		if instance == nil {
			return nil
		}
		return resource.RetryableError(fmt.Errorf("cfs snapshot status is %s, retry...", *instance.Status))
	})
	if err != nil {
		return err
	}

	return nil
}
