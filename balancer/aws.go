package balancer

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func GetASGInstances(asgName, region string) []string {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	asgSvc := autoscaling.NewFromConfig(cfg)
	asgInput := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []string{asgName},
	}

	asgResult, err := asgSvc.DescribeAutoScalingGroups(context.TODO(), asgInput)
	if err != nil {
		log.Fatalf("unable to describe ASG, %v", err)
	}

	var instanceIds []string
	for _, group := range asgResult.AutoScalingGroups {
		for _, instance := range group.Instances {
			instanceIds = append(instanceIds, *instance.InstanceId)
		}
	}

	ec2Svc := ec2.NewFromConfig(cfg)
	ec2Input := &ec2.DescribeInstancesInput{
		InstanceIds: instanceIds,
	}

	ec2Result, err := ec2Svc.DescribeInstances(context.TODO(), ec2Input)
	if err != nil {
		log.Fatalf("unable to describe EC2 instances, %v", err)
	}

	var instances []string
	for _, reservation := range ec2Result.Reservations {
		for _, instance := range reservation.Instances {
			instances = append(instances, *instance.PrivateIpAddress)
		}
	}

	return instances
}

func GetInstancesByTag(tagKey, tagValue, region string) []string {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	ec2Svc := ec2.NewFromConfig(cfg)
	input := &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:" + tagKey),
				Values: []string{tagValue},
			},
		},
	}

	result, err := ec2Svc.DescribeInstances(context.TODO(), input)
	if err != nil {
		log.Fatalf("unable to describe EC2 instances, %v", err)
	}

	var instances []string
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			instances = append(instances, *instance.PrivateIpAddress)
		}
	}

	return instances
}
