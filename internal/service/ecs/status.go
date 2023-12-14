// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	serviceStatusInactive = "INACTIVE"
	serviceStatusActive   = "ACTIVE"
	serviceStatusDraining = "DRAINING"
	// Non-standard statuses for statusServiceWaitForStable()
	serviceStatusPending = "tfPENDING"
	serviceStatusStable  = "tfSTABLE"

	taskSetStatusActive   = "ACTIVE"
	taskSetStatusDraining = "DRAINING"
	taskSetStatusPrimary  = "PRIMARY"
)

func statusCapacityProvider(ctx context.Context, conn *ecs.Client, arn, partition string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findCapacityProviderByARN(ctx, conn, arn, partition)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusCapacityProviderUpdate(ctx context.Context, conn *ecs.Client, arn, partition string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findCapacityProviderByARN(ctx, conn, arn, partition)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.UpdateStatus), nil
	}
}

func statusServiceNoTags(ctx context.Context, conn *ecs.Client, id, cluster, partition string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		service, err := findServiceNoTagsByID(ctx, conn, id, cluster, partition)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return service, aws.ToString(service.Status), err
	}
}

func statusServiceWaitForStable(ctx context.Context, conn *ecs.Client, id, cluster, partition string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		serviceRaw, status, err := statusServiceNoTags(ctx, conn, id, cluster, partition)()
		if err != nil {
			return nil, "", err
		}

		if status != serviceStatusActive {
			return serviceRaw, status, nil
		}

		service := serviceRaw.(*types.Service)

		if d, dc, rc := len(service.Deployments),
			service.DesiredCount,
			service.RunningCount; d == 1 && dc == rc {
			status = serviceStatusStable
		} else {
			status = serviceStatusPending
		}

		return service, status, nil
	}
}

func stabilityStatusTaskSet(ctx context.Context, conn *ecs.Client, taskSetID, service, cluster string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &ecs.DescribeTaskSetsInput{
			Cluster:  aws.String(cluster),
			Service:  aws.String(service),
			TaskSets: []string{taskSetID},
		}

		output, err := conn.DescribeTaskSets(ctx, input)

		if err != nil {
			return nil, "", err
		}

		if output == nil || len(output.TaskSets) == 0 {
			return nil, "", nil
		}

		return output.TaskSets[0], string(output.TaskSets[0].StabilityStatus), nil
	}
}

func statusTaskSet(ctx context.Context, conn *ecs.Client, taskSetID, service, cluster string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &ecs.DescribeTaskSetsInput{
			Cluster:  aws.String(cluster),
			Service:  aws.String(service),
			TaskSets: []string{taskSetID},
		}

		output, err := conn.DescribeTaskSets(ctx, input)

		if err != nil {
			return nil, "", err
		}

		if output == nil || len(output.TaskSets) == 0 {
			return nil, "", nil
		}

		return output.TaskSets[0], aws.ToString(output.TaskSets[0].Status), nil
	}
}
