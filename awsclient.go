package main

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type SgModel struct {
	ec2.SecurityGroup
	SGRequirement Sg
}

func initAws() *ec2.EC2 {
	sess, err := session.NewSession(&aws.Config{Region: aws.String("eu-west-2"), Endpoint: aws.String("https://fcu.eu-west-2.outscale.com")})

	if err != nil {
		fmt.Println(err)
	}

	return ec2.New(sess)
}

func syncSgIps(myIp string, svc *ec2.EC2, sgs []Sg) {
	var wg sync.WaitGroup
	for _, sg := range getSecurityGroups(sgs, svc) {
		wg.Add(1)
		go doSync(svc, sg, myIp, &wg)
	}
	wg.Wait()
}

func doSync(svc *ec2.EC2, sg SgModel, ip string, wg *sync.WaitGroup) {
	defer wg.Done()

	for _, perm := range sg.IpPermissions {
		// TODO: filter in query ( seems not working with outscale api )
		if inPortRange(perm, sg.SGRequirement.Port, sg.SGRequirement.Port) && !inIpRanges(ip, perm.IpRanges) {
			fmt.Printf("Syncing %s\n", *sg.GroupId)
			authorizeSg(svc, sg, ip)
		}
	}
}

func authorizeSg(svc *ec2.EC2, sg SgModel, ip string) {
	input := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: sg.GroupId,
		IpPermissions: []*ec2.IpPermission{
			{
				FromPort:   &sg.SGRequirement.Port,
				IpProtocol: aws.String("tcp"),
				IpRanges: []*ec2.IpRange{
					{
						CidrIp:      aws.String(ip),
						Description: &sg.SGRequirement.Comment,
					},
				},
				ToPort: &sg.SGRequirement.Port,
			},
		},
	}

	_, err := svc.AuthorizeSecurityGroupIngress(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}
}

func getSecurityGroups(sgs []Sg, svc *ec2.EC2) []SgModel {
	var groupIds []string

	for _, sg := range sgs {
		groupIds = append(groupIds, sg.Id)
	}

	result, err := svc.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
		GroupIds: aws.StringSlice(groupIds),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "InvalidGroupId.Malformed":
				fallthrough
			case "InvalidGroup.NotFound":
				Errorf("%s.", aerr.Message())
			}
		}
		Errorf("Unable to get descriptions for security groups, %v", err)
	}

	return mapSgtoData(result.SecurityGroups, sgs)
}

func inIpRanges(ip string, ipRanges []*ec2.IpRange) bool {
	for _, r := range ipRanges {
		if strings.Split(*r.CidrIp, "/")[0] == ip {
			return true
		}
	}
	return false
}

func inPortRange(perm *ec2.IpPermission, start int64, end int64) bool {
	return *perm.FromPort >= start && *perm.ToPort <= end
}

func mapSgtoData(awsSgs []*ec2.SecurityGroup, sgs []Sg) []SgModel {
	var sgModels []SgModel

	for _, sg := range awsSgs {
		for i := 0; i < len(sgs); i++ {
			if sgs[i].Id == *sg.GroupId {
				sgModels = append(sgModels, SgModel{*sg, sgs[i]})
				continue
			}
		}
	}

	return sgModels
}

func Errorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
}
