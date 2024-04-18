// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_iam_user_ssh_key", name="User SSH Key")
func resourceUserSSHKey() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserSSHKeyCreate,
		ReadWithoutTimeout:   resourceUserSSHKeyRead,
		UpdateWithoutTimeout: resourceUserSSHKeyUpdate,
		DeleteWithoutTimeout: resourceUserSSHKeyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceUserSSHKeyImport,
		},

		Schema: map[string]*schema.Schema{
			"encoding": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.EncodingType](),
			},
			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if d.Get("encoding").(string) == "SSH" {
						old = cleanSSHKey(old)
						new = cleanSSHKey(new)
					}
					return strings.Trim(old, "\n") == strings.Trim(new, "\n")
				},
			},
			"ssh_public_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"username": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceUserSSHKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	username := d.Get("username").(string)
	input := &iam.UploadSSHPublicKeyInput{
		SSHPublicKeyBody: aws.String(d.Get("public_key").(string)),
		UserName:         aws.String(username),
	}

	output, err := conn.UploadSSHPublicKey(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "uploading IAM User SSH Key (%s): %s", username, err)
	}

	d.SetId(aws.ToString(output.SSHPublicKey.SSHPublicKeyId))

	_, err = tfresource.RetryWhenNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return findSSHPublicKeyByThreePartKey(ctx, conn, d.Id(), d.Get("encoding").(string), username)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IAM User SSH Key (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("status"); ok {
		input := &iam.UpdateSSHPublicKeyInput{
			SSHPublicKeyId: aws.String(d.Id()),
			Status:         awstypes.StatusType(v.(string)),
			UserName:       aws.String(username),
		}

		_, err := conn.UpdateSSHPublicKey(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM User SSH Key (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceUserSSHKeyRead(ctx, d, meta)...)
}

func resourceUserSSHKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	encoding := d.Get("encoding").(string)
	key, err := findSSHPublicKeyByThreePartKey(ctx, conn, d.Id(), encoding, d.Get("username").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM User SSH Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM User SSH Key (%s): %s", d.Id(), err)
	}

	d.Set("fingerprint", key.Fingerprint)
	publicKey := aws.ToString(key.SSHPublicKeyBody)
	if encoding == string(awstypes.EncodingTypeSsh) {
		publicKey = cleanSSHKey(publicKey)
	}
	d.Set("public_key", publicKey)
	d.Set("ssh_public_key_id", key.SSHPublicKeyId)
	d.Set("status", key.Status)

	return diags
}

func resourceUserSSHKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	input := &iam.UpdateSSHPublicKeyInput{
		SSHPublicKeyId: aws.String(d.Id()),
		Status:         awstypes.StatusType(d.Get("status").(string)),
		UserName:       aws.String(d.Get("username").(string)),
	}

	_, err := conn.UpdateSSHPublicKey(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating IAM User SSH Key (%s): %s", d.Id(), err)
	}

	return append(diags, resourceUserSSHKeyRead(ctx, d, meta)...)
}

func resourceUserSSHKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	log.Printf("[DEBUG] Deleting IAM User SSH Key: %s", d.Id())
	_, err := conn.DeleteSSHPublicKey(ctx, &iam.DeleteSSHPublicKeyInput{
		SSHPublicKeyId: aws.String(d.Id()),
		UserName:       aws.String(d.Get("username").(string)),
	})

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM User SSH Key (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceUserSSHKeyImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.SplitN(d.Id(), ":", 3)

	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), UserName:SSHPublicKeyId:Encoding", d.Id())
	}

	username := idParts[0]
	sshPublicKeyId := idParts[1]
	encoding := idParts[2]

	d.Set("username", username)
	d.Set("ssh_public_key_id", sshPublicKeyId)
	d.Set("encoding", encoding)
	d.SetId(sshPublicKeyId)

	return []*schema.ResourceData{d}, nil
}

func findSSHPublicKeyByThreePartKey(ctx context.Context, conn *iam.Client, id, encoding, username string) (*awstypes.SSHPublicKey, error) {
	input := &iam.GetSSHPublicKeyInput{
		Encoding:       awstypes.EncodingType(encoding),
		SSHPublicKeyId: aws.String(id),
		UserName:       aws.String(username),
	}

	output, err := conn.GetSSHPublicKey(ctx, input)

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.SSHPublicKey == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.SSHPublicKey, nil
}

func cleanSSHKey(key string) string {
	// Remove comments from SSH Keys
	// Comments are anything after "ssh-rsa XXXX" where XXXX is the key.
	parts := strings.Split(key, " ")
	if len(parts) > 2 {
		parts = parts[0:2]
	}
	return strings.Join(parts, " ")
}
