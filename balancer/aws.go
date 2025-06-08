package balancer

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

const (
	maxRetries = 3
	retryDelay = time.Second
)

func GetASGInstancesWithRetry(asgName, region string) ([]string, error) {
	var instances []string
	var err error

	for i := 0; i < maxRetries; i++ {
		instances, err = GetASGInstances(asgName, region)
		if err == nil {
			return instances, nil
		}
		log.Printf("Attempt %d failed to get ASG instances: %v", i+1, err)
		time.Sleep(retryDelay * time.Duration(i+1))
	}

	return nil, err
}

func GetInstancesByTagWithRetry(tagKey, tagValue, region string) ([]string, error) {
	var instances []string
	var err error

	for i := 0; i < maxRetries; i++ {
		instances, err = GetInstancesByTag(tagKey, tagValue, region)
		if err == nil {
			return instances, nil
		}
		log.Printf("Attempt %d failed to get instances by tag: %v", i+1, err)
		time.Sleep(retryDelay * time.Duration(i+1))
	}

	return nil, err
}

func GetASGInstances(asgName, region string) ([]string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %v", err)
	}

	asgSvc := autoscaling.NewFromConfig(cfg)
	asgInput := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []string{asgName},
	}

	asgResult, err := asgSvc.DescribeAutoScalingGroups(context.TODO(), asgInput)
	if err != nil {
		return nil, fmt.Errorf("unable to describe ASG: %v", err)
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
		return nil, fmt.Errorf("unable to describe EC2 instances: %v", err)
	}

	var instances []string
	for _, reservation := range ec2Result.Reservations {
		for _, instance := range reservation.Instances {
			instances = append(instances, *instance.PrivateIpAddress)
		}
	}

	return instances, nil
}

func GetInstancesByTag(tagKey, tagValue, region string) ([]string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %v", err)
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
		return nil, fmt.Errorf("unable to describe EC2 instances: %v", err)
	}

	var instances []string
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			instances = append(instances, *instance.PrivateIpAddress)
		}
	}

	return instances, nil
}
