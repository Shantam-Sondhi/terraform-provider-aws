---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_endpoint_service_private_dns_verification"
description: |-
  Terraform resource for managing an AWS VPC (Virtual Private Cloud) Endpoint Service Private DNS Verification.
---
# Resource: aws_vpc_endpoint_service_private_dns_verification

Terraform resource for managing an AWS VPC (Virtual Private Cloud) Endpoint Service Private DNS Verification.

~> Destruction of this resource will not stop the verification process, only remove the resource from state.

## Example Usage

### Basic Usage

```terraform
resource "aws_vpc_endpoint_service_private_dns_verification" "example" {
  service_id = aws_vpc_endpoint_service.example.id
}
```

## Argument Reference

The following arguments are required:

* `service_id` - (Required) ID of the endpoint service.

## Attribute Reference

This resource exports no additional attributes.

## Import

You cannot import this resource.
