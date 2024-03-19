package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/elbv2"
)

func getAutoScalingGroupName(svc *autoscaling.AutoScaling, instanceID string) (string, error) {
	input := &autoscaling.DescribeAutoScalingInstancesInput{
		InstanceIds: []*string{aws.String(instanceID)},
	}

	result, err := svc.DescribeAutoScalingInstances(input)
	if err != nil {
		return "", err
	}

	if len(result.AutoScalingInstances) == 0 {
		return "", fmt.Errorf("Instance %s not found, or does not belong to an Auto Scaling Group", instanceID)
	}

	return *result.AutoScalingInstances[0].AutoScalingGroupName, nil
}

func getTargetGroupARNs(svc *autoscaling.AutoScaling, autoScalingGroupName string) ([]string, error) {
	input := &autoscaling.DescribeLoadBalancerTargetGroupsInput{
		AutoScalingGroupName: aws.String(autoScalingGroupName),
	}

	result, err := svc.DescribeLoadBalancerTargetGroups(input)
	if err != nil {
		return nil, err
	}

	var arns []string
	for _, tg := range result.LoadBalancerTargetGroups {
		arns = append(arns, *tg.LoadBalancerTargetGroupARN)
	}
	return arns, nil
}

func waitUntilInService(svc *elbv2.ELBV2, instanceID, targetGroupARN string) {
	input := &elbv2.DescribeTargetHealthInput{
		TargetGroupArn: aws.String(targetGroupARN),
		Targets: []*elbv2.TargetDescription{
			{
				Id: aws.String(instanceID),
			},
		},
	}

	err := svc.WaitUntilTargetInService(input)
	if err != nil {
		log.Fatalf("Error waiting for instance %s to be in service: %v", instanceID, err)
	}
}

func main() {
	instanceID := flag.String("instance-id", "", "The ID of the instance")
	flag.Parse()

	if *instanceID == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	log.Println("Waiting until instance is healthy in its ELBs")

	sess := session.Must(session.NewSession())
	asgSvc := autoscaling.New(sess)
	elbv2Svc := elbv2.New(sess)

	asgName, err := getAutoScalingGroupName(asgSvc, *instanceID)
	if err != nil {
		log.Fatalf("Error getting ASG name: %v", err)
	}
	log.Printf("Found ASG: %s", asgName)

	targetGroupARNs, err := getTargetGroupARNs(asgSvc, asgName)
	if err != nil {
		log.Fatalf("Error getting target group ARNs: %v", err)
	}
	log.Printf("ASG configures %d target groups", len(targetGroupARNs))

	for _, targetGroupARN := range targetGroupARNs {
		targetGroupName := strings.Split(targetGroupARN, "/")[1]
		log.Printf("Waiting for instance to register healthy in %s", targetGroupName)
		waitUntilInService(elbv2Svc, *instanceID, targetGroupARN)
	}

	log.Println("Instance showing as healthy in all TGs")
}
