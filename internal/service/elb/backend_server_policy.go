package elb

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceBackendServerPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBackendServerPolicySet,
		ReadWithoutTimeout:   resourceBackendServerPolicyRead,
		UpdateWithoutTimeout: resourceBackendServerPolicySet,
		DeleteWithoutTimeout: resourceBackendServerPolicyDelete,

		Schema: map[string]*schema.Schema{
			"instance_port": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"load_balancer_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"policy_names": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
		},
	}
}

func resourceBackendServerPolicySet(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBConn()

	instancePort := d.Get("instance_port").(int)
	lbName := d.Get("load_balancer_name").(string)
	id := BackendServerPolicyCreateResourceID(lbName, instancePort)
	input := &elb.SetLoadBalancerPoliciesForBackendServerInput{
		InstancePort:     aws.Int64(int64(instancePort)),
		LoadBalancerName: aws.String(lbName),
	}

	if v, ok := d.GetOk("policy_names"); ok && v.(*schema.Set).Len() > 0 {
		input.PolicyNames = flex.ExpandStringSet(v.(*schema.Set))
	}

	_, err := conn.SetLoadBalancerPoliciesForBackendServerWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ELB Classic Backend Server Policy (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceBackendServerPolicyRead(ctx, d, meta)...)
}

func resourceBackendServerPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBConn()

	lbName, instancePort, err := BackendServerPolicyParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing resource ID: %s", err)
	}

	policyNames, err := FindLoadBalancerBackendServerPolicyByTwoPartKey(ctx, conn, lbName, instancePort)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ELB Classic Backend Server Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELB Classic Backend Server Policy (%s): %s", d.Id(), err)
	}

	d.Set("instance_port", instancePort)
	d.Set("load_balancer_name", lbName)
	d.Set("policy_names", policyNames)

	return diags
}

func resourceBackendServerPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBConn()

	lbName, instancePort, err := BackendServerPolicyParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing resource ID: %s", err)
	}

	input := &elb.SetLoadBalancerPoliciesForBackendServerInput{
		InstancePort:     aws.Int64(int64(instancePort)),
		LoadBalancerName: aws.String(lbName),
		PolicyNames:      aws.StringSlice([]string{}),
	}

	log.Printf("[DEBUG] Deleting ELB Classic Backend Server Policy: %s", d.Id())
	_, err = conn.SetLoadBalancerPoliciesForBackendServerWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ELB Classic Backend Server Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func FindLoadBalancerBackendServerPolicyByTwoPartKey(ctx context.Context, conn *elb.ELB, lbName string, instancePort int) ([]string, error) {
	lb, err := FindLoadBalancerByName(ctx, conn, lbName)

	if err != nil {
		return nil, err
	}

	var policyNames []string

	for _, v := range lb.BackendServerDescriptions {
		if v == nil {
			continue
		}

		if aws.Int64Value(v.InstancePort) != int64(instancePort) {
			continue
		}

		policyNames = append(policyNames, aws.StringValueSlice(v.PolicyNames)...)
	}

	return policyNames, nil
}

const backendServerPolicyResourceIDSeparator = ":"

func BackendServerPolicyCreateResourceID(lbName string, instancePort int) string {
	parts := []string{lbName, strconv.Itoa(instancePort)}
	id := strings.Join(parts, backendServerPolicyResourceIDSeparator)

	return id
}

func BackendServerPolicyParseResourceID(id string) (string, int, error) {
	parts := strings.Split(id, backendServerPolicyResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		v, err := strconv.Atoi(parts[1])

		if err != nil {
			return "", 0, err
		}

		return parts[0], v, nil
	}

	return "", 0, fmt.Errorf("unexpected format for ID (%[1]s), expected LBNAME%[2]sINSTANCEPORT", id, backendServerPolicyResourceIDSeparator)
}
