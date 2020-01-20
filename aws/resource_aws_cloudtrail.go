package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsCloudTrail() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudTrailCreate,
		Read:   resourceAwsCloudTrailRead,
		Update: resourceAwsCloudTrailUpdate,
		Delete: resourceAwsCloudTrailDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"enable_logging": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"s3_bucket_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"s3_key_prefix": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"cloud_watch_logs_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},
			"cloud_watch_logs_group_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},
			"include_global_service_events": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"is_multi_region_trail": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"is_organization_trail": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"sns_topic_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"enable_log_file_validation": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},
			"event_selector": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 5,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"read_write_type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  cloudtrail.ReadWriteTypeAll,
							ValidateFunc: validation.StringInSlice([]string{
								cloudtrail.ReadWriteTypeAll,
								cloudtrail.ReadWriteTypeReadOnly,
								cloudtrail.ReadWriteTypeWriteOnly,
							}, false),
						},

						"include_management_events": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},

						"data_resource": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"AWS::S3::Object", "AWS::Lambda::Function"}, false),
									},
									"values": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 250,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validateArn,
										},
									},
								},
							},
						},
						"exclude_management_event_sources": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"home_region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
			"sns_topic_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsCloudTrailCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudtrailconn
	tags := keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().CloudtrailTags()

	input := cloudtrail.CreateTrailInput{
		Name:         aws.String(d.Get("name").(string)),
		S3BucketName: aws.String(d.Get("s3_bucket_name").(string)),
	}

	if len(tags) > 0 {
		input.TagsList = tags
	}

	if v, ok := d.GetOk("cloud_watch_logs_group_arn"); ok {
		input.CloudWatchLogsLogGroupArn = aws.String(v.(string))
	}
	if v, ok := d.GetOk("cloud_watch_logs_role_arn"); ok {
		input.CloudWatchLogsRoleArn = aws.String(v.(string))
	}
	if v, ok := d.GetOkExists("include_global_service_events"); ok {
		input.IncludeGlobalServiceEvents = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk("is_multi_region_trail"); ok {
		input.IsMultiRegionTrail = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk("is_organization_trail"); ok {
		input.IsOrganizationTrail = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk("enable_log_file_validation"); ok {
		input.EnableLogFileValidation = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("s3_key_prefix"); ok {
		input.S3KeyPrefix = aws.String(v.(string))
	}
	if v, ok := d.GetOk("sns_topic_name"); ok {
		input.SnsTopicName = aws.String(v.(string))
	}

	var t *cloudtrail.CreateTrailOutput
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error
		t, err = conn.CreateTrail(&input)
		if err != nil {
			if isAWSErr(err, cloudtrail.ErrCodeInvalidCloudWatchLogsRoleArnException, "Access denied.") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, cloudtrail.ErrCodeInvalidCloudWatchLogsLogGroupArnException, "Access denied.") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		t, err = conn.CreateTrail(&input)
	}
	if err != nil {
		return fmt.Errorf("Error creating CloudTrail: %s", err)
	}

	log.Printf("[DEBUG] CloudTrail created: %s", t)

	d.SetId(*t.Name)

	// AWS CloudTrail sets newly-created trails to false.
	if v, ok := d.GetOk("enable_logging"); ok && v.(bool) {
		err := cloudTrailSetLogging(conn, v.(bool), d.Id())
		if err != nil {
			return err
		}
	}

	// Event Selectors
	if _, ok := d.GetOk("event_selector"); ok {
		if err := cloudTrailSetEventSelectors(conn, d); err != nil {
			return err
		}
	}

	return resourceAwsCloudTrailRead(d, meta)
}

func resourceAwsCloudTrailRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudtrailconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := cloudtrail.DescribeTrailsInput{
		TrailNameList: []*string{
			aws.String(d.Id()),
		},
	}
	resp, err := conn.DescribeTrails(&input)
	if err != nil {
		return err
	}

	// CloudTrail does not return a NotFound error in the event that the Trail
	// you're looking for is not found. Instead, it's simply not in the list.
	var trail *cloudtrail.Trail
	for _, c := range resp.TrailList {
		if d.Id() == *c.Name {
			trail = c
		}
	}

	if trail == nil {
		log.Printf("[WARN] CloudTrail (%s) not found", d.Id())
		d.SetId("")
		return nil
	}

	log.Printf("[DEBUG] CloudTrail received: %s", trail)

	d.Set("name", trail.Name)
	d.Set("s3_bucket_name", trail.S3BucketName)
	d.Set("s3_key_prefix", trail.S3KeyPrefix)
	d.Set("cloud_watch_logs_role_arn", trail.CloudWatchLogsRoleArn)
	d.Set("cloud_watch_logs_group_arn", trail.CloudWatchLogsLogGroupArn)
	d.Set("include_global_service_events", trail.IncludeGlobalServiceEvents)
	d.Set("is_multi_region_trail", trail.IsMultiRegionTrail)
	d.Set("is_organization_trail", trail.IsOrganizationTrail)
	d.Set("sns_topic_name", trail.SnsTopicName)
	d.Set("sns_topic_arn", trail.SnsTopicARN)
	d.Set("enable_log_file_validation", trail.LogFileValidationEnabled)

	// TODO: Make it possible to use KMS Key names, not just ARNs
	// In order to test it properly this PR needs to be merged 1st:
	// https://github.com/hashicorp/terraform/pull/3928
	d.Set("kms_key_id", trail.KmsKeyId)

	d.Set("arn", trail.TrailARN)
	d.Set("home_region", trail.HomeRegion)

	tags, err := keyvaluetags.CloudtrailListTags(conn, *trail.TrailARN)

	if err != nil {
		return fmt.Errorf("error listing tags for Cloudtrail (%s): %s", *trail.TrailARN, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	logstatus, err := cloudTrailGetLoggingStatus(conn, trail.Name)
	if err != nil {
		return err
	}
	d.Set("enable_logging", logstatus)

	// Get EventSelectors
	eventSelectorsOut, err := conn.GetEventSelectors(&cloudtrail.GetEventSelectorsInput{
		TrailName: aws.String(d.Id()),
	})
	if err != nil {
		return err
	}

	if err := d.Set("event_selector", flattenAwsCloudTrailEventSelector(eventSelectorsOut.EventSelectors)); err != nil {
		return err
	}

	return nil
}

func resourceAwsCloudTrailUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudtrailconn

	input := cloudtrail.UpdateTrailInput{
		Name: aws.String(d.Id()),
	}

	if d.HasChange("s3_bucket_name") {
		input.S3BucketName = aws.String(d.Get("s3_bucket_name").(string))
	}
	if d.HasChange("s3_key_prefix") {
		input.S3KeyPrefix = aws.String(d.Get("s3_key_prefix").(string))
	}
	if d.HasChanges("cloud_watch_logs_role_arn", "cloud_watch_logs_group_arn") {
		// Both of these need to be provided together
		// in the update call otherwise API complains
		input.CloudWatchLogsRoleArn = aws.String(d.Get("cloud_watch_logs_role_arn").(string))
		input.CloudWatchLogsLogGroupArn = aws.String(d.Get("cloud_watch_logs_group_arn").(string))
	}
	if d.HasChange("include_global_service_events") {
		input.IncludeGlobalServiceEvents = aws.Bool(d.Get("include_global_service_events").(bool))
	}
	if d.HasChange("is_multi_region_trail") {
		input.IsMultiRegionTrail = aws.Bool(d.Get("is_multi_region_trail").(bool))
	}
	if d.HasChange("is_organization_trail") {
		input.IsOrganizationTrail = aws.Bool(d.Get("is_organization_trail").(bool))
	}
	if d.HasChange("enable_log_file_validation") {
		input.EnableLogFileValidation = aws.Bool(d.Get("enable_log_file_validation").(bool))
	}
	if d.HasChange("kms_key_id") {
		input.KmsKeyId = aws.String(d.Get("kms_key_id").(string))
	}
	if d.HasChange("sns_topic_name") {
		input.SnsTopicName = aws.String(d.Get("sns_topic_name").(string))
	}

	log.Printf("[DEBUG] Updating CloudTrail: %s", input)
	var t *cloudtrail.UpdateTrailOutput
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error
		t, err = conn.UpdateTrail(&input)
		if err != nil {
			if isAWSErr(err, cloudtrail.ErrCodeInvalidCloudWatchLogsRoleArnException, "Access denied.") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, cloudtrail.ErrCodeInvalidCloudWatchLogsLogGroupArnException, "Access denied.") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		t, err = conn.UpdateTrail(&input)
	}
	if err != nil {
		return fmt.Errorf("Error updating CloudTrail: %s", err)
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.CloudtrailUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating ECR Repository (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	if d.HasChange("enable_logging") {
		log.Printf("[DEBUG] Updating logging on CloudTrail: %s", input)
		err := cloudTrailSetLogging(conn, d.Get("enable_logging").(bool), *input.Name)
		if err != nil {
			return err
		}
	}

	if !d.IsNewResource() && d.HasChange("event_selector") {
		log.Printf("[DEBUG] Updating event selector on CloudTrail: %s", input)
		if err := cloudTrailSetEventSelectors(conn, d); err != nil {
			return err
		}
	}

	log.Printf("[DEBUG] CloudTrail updated: %s", t)

	return resourceAwsCloudTrailRead(d, meta)
}

func resourceAwsCloudTrailDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudtrailconn

	log.Printf("[DEBUG] Deleting CloudTrail: %q", d.Id())
	_, err := conn.DeleteTrail(&cloudtrail.DeleteTrailInput{
		Name: aws.String(d.Id()),
	})

	return err
}

func cloudTrailGetLoggingStatus(conn *cloudtrail.CloudTrail, id *string) (bool, error) {
	GetTrailStatusOpts := &cloudtrail.GetTrailStatusInput{
		Name: id,
	}
	resp, err := conn.GetTrailStatus(GetTrailStatusOpts)
	if err != nil {
		return false, fmt.Errorf("Error retrieving logging status of CloudTrail (%s): %s", *id, err)
	}

	return *resp.IsLogging, err
}

func cloudTrailSetLogging(conn *cloudtrail.CloudTrail, enabled bool, id string) error {
	if enabled {
		log.Printf(
			"[DEBUG] Starting logging on CloudTrail (%s)",
			id)
		StartLoggingOpts := &cloudtrail.StartLoggingInput{
			Name: aws.String(id),
		}
		if _, err := conn.StartLogging(StartLoggingOpts); err != nil {
			return fmt.Errorf(
				"Error starting logging on CloudTrail (%s): %s",
				id, err)
		}
	} else {
		log.Printf(
			"[DEBUG] Stopping logging on CloudTrail (%s)",
			id)
		StopLoggingOpts := &cloudtrail.StopLoggingInput{
			Name: aws.String(id),
		}
		if _, err := conn.StopLogging(StopLoggingOpts); err != nil {
			return fmt.Errorf(
				"Error stopping logging on CloudTrail (%s): %s",
				id, err)
		}
	}

	return nil
}

func cloudTrailSetEventSelectors(conn *cloudtrail.CloudTrail, d *schema.ResourceData) error {
	input := &cloudtrail.PutEventSelectorsInput{
		TrailName: aws.String(d.Id()),
	}

	eventSelectors := expandAwsCloudTrailEventSelector(d.Get("event_selector").([]interface{}))
	// If no defined selectors revert to the single default selector
	if len(eventSelectors) == 0 {
		es := &cloudtrail.EventSelector{
			IncludeManagementEvents: aws.Bool(true),
			ReadWriteType:           aws.String("All"),
			DataResources:           make([]*cloudtrail.DataResource, 0),
		}
		eventSelectors = append(eventSelectors, es)
	}
	input.EventSelectors = eventSelectors

	if err := input.Validate(); err != nil {
		return fmt.Errorf("Error validate CloudTrail (%s): %s", d.Id(), err)
	}

	_, err := conn.PutEventSelectors(input)
	if err != nil {
		return fmt.Errorf("Error set event selector on CloudTrail (%s): %s", d.Id(), err)
	}

	return nil
}

func expandAwsCloudTrailEventSelector(configured []interface{}) []*cloudtrail.EventSelector {
	eventSelectors := make([]*cloudtrail.EventSelector, 0, len(configured))

	for _, raw := range configured {
		data := raw.(map[string]interface{})
		dataResources := expandAwsCloudTrailEventSelectorDataResource(data["data_resource"].([]interface{}))

		includeManagementEvents := data["include_management_events"].(bool)
		es := &cloudtrail.EventSelector{
			IncludeManagementEvents: aws.Bool(includeManagementEvents),
			ReadWriteType:           aws.String(data["read_write_type"].(string)),
			DataResources:           dataResources,
		}
		if data["exclude_management_event_sources"] != nil && includeManagementEvents {
			es.ExcludeManagementEventSources = expandStringList(data["exclude_management_event_sources"].([]interface{}))
		}

		eventSelectors = append(eventSelectors, es)
	}

	return eventSelectors
}

func expandAwsCloudTrailEventSelectorDataResource(configured []interface{}) []*cloudtrail.DataResource {
	dataResources := make([]*cloudtrail.DataResource, 0, len(configured))

	for _, raw := range configured {
		data := raw.(map[string]interface{})

		values := make([]*string, len(data["values"].([]interface{})))
		for i, vv := range data["values"].([]interface{}) {
			str := vv.(string)
			values[i] = aws.String(str)
		}

		dataResource := &cloudtrail.DataResource{
			Type:   aws.String(data["type"].(string)),
			Values: values,
		}

		dataResources = append(dataResources, dataResource)
	}

	return dataResources
}

func flattenAwsCloudTrailEventSelector(configured []*cloudtrail.EventSelector) []map[string]interface{} {
	eventSelectors := make([]map[string]interface{}, 0, len(configured))

	// Prevent default configurations shows differences
	if len(configured) == 1 && len(configured[0].DataResources) == 0 && aws.StringValue(configured[0].ReadWriteType) == "All" {
		return eventSelectors
	}

	for _, raw := range configured {
		item := make(map[string]interface{})
		item["read_write_type"] = *raw.ReadWriteType
		item["include_management_events"] = *raw.IncludeManagementEvents
		item["data_resource"] = flattenAwsCloudTrailEventSelectorDataResource(raw.DataResources)
		item["exclude_management_event_sources"] = flattenStringList(raw.ExcludeManagementEventSources)

		eventSelectors = append(eventSelectors, item)
	}

	return eventSelectors
}

func flattenAwsCloudTrailEventSelectorDataResource(configured []*cloudtrail.DataResource) []map[string]interface{} {
	dataResources := make([]map[string]interface{}, 0, len(configured))

	for _, raw := range configured {
		item := make(map[string]interface{})
		item["type"] = *raw.Type
		item["values"] = flattenStringList(raw.Values)

		dataResources = append(dataResources, item)
	}

	return dataResources
}
