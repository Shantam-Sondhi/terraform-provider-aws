// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ce

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	awstypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ce_anomaly_subscription", name="Anomaly Subscription")
// @Tags(identifierAttribute="id")
func resourceAnomalySubscription() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAnomalySubscriptionCreate,
		ReadWithoutTimeout:   resourceAnomalySubscriptionRead,
		UpdateWithoutTimeout: resourceAnomalySubscriptionUpdate,
		DeleteWithoutTimeout: resourceAnomalySubscriptionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"frequency": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.AnomalySubscriptionFrequency](),
			},
			"monitor_arn_list": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 1024),
					validation.StringMatch(regexache.MustCompile(`[\\S\\s]*`), "Must be a valid Anomaly Subscription Name matching expression: [\\S\\s]*")),
			},
			"subscriber": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.SubscriberType](),
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"threshold_expression": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Computed: true,
				Optional: true,
				Elem:     elemExpression(),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAnomalySubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CEClient(ctx)

	name := d.Get("name").(string)
	input := &costexplorer.CreateAnomalySubscriptionInput{
		AnomalySubscription: &awstypes.AnomalySubscription{
			Frequency:        awstypes.AnomalySubscriptionFrequency(d.Get("frequency").(string)),
			MonitorArnList:   flex.ExpandStringValueList(d.Get("monitor_arn_list").([]interface{})),
			Subscribers:      expandSubscribers(d.Get("subscriber").(*schema.Set).List()),
			SubscriptionName: aws.String(name),
		},
		ResourceTags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk("account_id"); ok {
		input.AnomalySubscription.AccountId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("threshold_expression"); ok {
		input.AnomalySubscription.ThresholdExpression = expandCostExpression(v.([]interface{})[0].(map[string]interface{}))
	}

	output, err := conn.CreateAnomalySubscription(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Cost Explorer Anomaly Subscription (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.SubscriptionArn))

	return append(diags, resourceAnomalySubscriptionRead(ctx, d, meta)...)
}

func resourceAnomalySubscriptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CEClient(ctx)

	subscription, err := findAnomalySubscriptionByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Cost Explorer Anomaly Subscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Cost Explorer Anomaly Subscription (%s): %s", d.Id(), err)
	}

	d.Set("account_id", subscription.AccountId)
	d.Set("arn", subscription.SubscriptionArn)
	d.Set("frequency", subscription.Frequency)
	d.Set("monitor_arn_list", subscription.MonitorArnList)
	d.Set("name", subscription.SubscriptionName)
	d.Set("subscriber", flattenSubscribers(subscription.Subscribers))
	if err := d.Set("threshold_expression", []interface{}{flattenCostCategoryRuleExpression(subscription.ThresholdExpression)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting threshold_expression: %s", err)
	}

	return diags
}

func resourceAnomalySubscriptionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CEClient(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &costexplorer.UpdateAnomalySubscriptionInput{
			SubscriptionArn: aws.String(d.Id()),
		}

		if d.HasChange("frequency") {
			input.Frequency = awstypes.AnomalySubscriptionFrequency(d.Get("frequency").(string))
		}

		if d.HasChange("monitor_arn_list") {
			input.MonitorArnList = flex.ExpandStringValueList(d.Get("monitor_arn_list").([]interface{}))
		}

		if d.HasChange("subscriber") {
			input.Subscribers = expandSubscribers(d.Get("subscriber").(*schema.Set).List())
		}

		if d.HasChange("threshold_expression") {
			input.ThresholdExpression = expandCostExpression(d.Get("threshold_expression").([]interface{})[0].(map[string]interface{}))
		}

		_, err := conn.UpdateAnomalySubscription(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Cost Explorer Anomaly Subscription (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceAnomalySubscriptionRead(ctx, d, meta)...)
}

func resourceAnomalySubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CEClient(ctx)

	log.Printf("[DEBUG] Deleting Cost Explorer Anomaly Subscription: %s", d.Id())
	_, err := conn.DeleteAnomalySubscription(ctx, &costexplorer.DeleteAnomalySubscriptionInput{
		SubscriptionArn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.UnknownSubscriptionException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Cost Explorer Anomaly Subscription (%s): %s", d.Id(), err)
	}

	return diags
}

func findAnomalySubscriptionByARN(ctx context.Context, conn *costexplorer.Client, arn string) (*awstypes.AnomalySubscription, error) {
	input := &costexplorer.GetAnomalySubscriptionsInput{
		SubscriptionArnList: []string{arn},
		MaxResults:          aws.Int32(1),
	}

	output, err := conn.GetAnomalySubscriptions(ctx, input)

	if errs.IsA[*awstypes.UnknownMonitorException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.AnomalySubscriptions) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return &output.AnomalySubscriptions[0], nil
}

func expandSubscribers(tfList []interface{}) []awstypes.Subscriber {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.Subscriber

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObjects = append(apiObjects, awstypes.Subscriber{
			Address: aws.String(tfMap["address"].(string)),
			Type:    awstypes.SubscriberType(tfMap["type"].(string)),
		})
	}

	return apiObjects
}

func flattenSubscribers(apiObjects []awstypes.Subscriber) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, map[string]interface{}{
			"address": aws.ToString(apiObject.Address),
			"type":    apiObject.Type,
		})
	}

	return tfList
}
