package template

import (
	"fmt"

	"github.com/awslabs/goformation/cloudformation"
	"github.com/awslabs/goformation/cloudformation/resources"
	"github.com/coinbase/fenrir/aws"
	"github.com/coinbase/fenrir/aws/lambda"
)

func ValidateAWSElasticLoadBalancingV2TargetGroup(
	projectName, configName, resourceName string,
	template *cloudformation.Template,
	lambdac aws.LambdaAPI,
	res *resources.AWSElasticLoadBalancingV2TargetGroup,
) error {
	res.Name = normalizeName("fenrir", projectName, configName, resourceName, 32)

	res.Tags = append(res.Tags, resources.Tag{Key: "ProjectName", Value: projectName})
	res.Tags = append(res.Tags, resources.Tag{Key: "ConfigName", Value: configName})
	res.Tags = append(res.Tags, resources.Tag{Key: "ServiceName", Value: resourceName})

	// Only allow lambda targets for now
	if res.TargetType != "lambda" {
		return resourceError(res, resourceName, "TargetGroup.TargetType must be lambda")
	}

	// Only allow lambdas created in this template.
	// In the future we'll want to allow targets to be any lambda with correct tags
	// (tags specifically allowing this project or all fenrir projects)
	for _, target := range res.Targets {
		args, err := decodeGetAtt(target.Id)
		if err != nil || len(args) != 2 || args[1] != "Arn" {
			lambda, err := lambda.FindFunction(lambdac, target.Id)
			if err != nil {
				return resourceError(res, resourceName, "TargetGroup.Targets.Id must be \"!GetAtt <lambdaName> Arn\" or a valid lambda ARN")
			}

			if err := hasCorrectTags(projectName, configName, convTagMap(lambda.Tags)); err != nil {
				return resourceError(res, resourceName, fmt.Sprintf("TargetGroup.Target %v", err.Error()))
			}
		}
	}

	return nil
}

func convTagMap(tags map[string]*string) map[string]string {
	newTags := map[string]string{}

	for k, v := range tags {
		if v == nil {
			newTags[k] = ""
		} else {
			newTags[k] = *v
		}
	}

	return newTags
}
