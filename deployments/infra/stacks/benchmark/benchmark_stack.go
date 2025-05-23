package benchmark

import (
	"fmt"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"github.com/trufnetwork/node/infra/config"
	goasset "github.com/trufnetwork/node/infra/lib/goasset"
)

// Main stack function
func BenchmarkStack(scope constructs.Construct, id string, props *awscdk.StackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, jsii.String(id), props)

	// Create S3 buckets for storing binaries and results
	binaryS3Asset := goasset.BundleDir(stack, "benchmark-binary", "../../internal/benchmark",
		func(o *goasset.Options) {
			o.IsTest = true
			o.OutName = "benchmark"
			// Add any specific BuildFlags or Platform if needed, otherwise defaults are used
			// o.Platform = "linux/amd64" // Default
			// o.BuildFlags = []string{"-tags=foo"}
		},
	)
	resultsBucket := createBucket(
		stack,
		"benchmark-results-"+strings.ToLower(*stack.StackName()),
	)

	// Define the EC2 instance types to be tested
	testedInstances := []awsec2.InstanceType{
		// we don't support micro. it either gets error or hangs during the tests
		//awsec2.InstanceType_Of(awsec2.InstanceClass_T3, awsec2.InstanceSize_MICRO),
		awsec2.InstanceType_Of(awsec2.InstanceClass_T3A, awsec2.InstanceSize_SMALL),
		awsec2.InstanceType_Of(awsec2.InstanceClass_T3A, awsec2.InstanceSize_MEDIUM),
		awsec2.InstanceType_Of(awsec2.InstanceClass_T3A, awsec2.InstanceSize_LARGE),
	}

	// default vpc
	defaultVPC := awsec2.Vpc_FromLookup(stack, jsii.String("VPC"), &awsec2.VpcLookupOptions{
		IsDefault: jsii.Bool(true),
	})

	securityGroup := awsec2.NewSecurityGroup(stack, jsii.String("benchmark-security-group"), &awsec2.SecurityGroupProps{
		Vpc: defaultVPC,
	})

	// permit 22 port for ssh
	securityGroup.AddIngressRule(
		awsec2.Peer_AnyIpv4(),
		awsec2.Port_Tcp(jsii.Number(22)),
		jsii.String("Allow SSH access"),
		jsii.Bool(true),
	)

	ec2InstanceRole := awsiam.NewRole(stack, jsii.String("EC2InstanceRole"), &awsiam.RoleProps{
		AssumedBy:   awsiam.NewServicePrincipal(jsii.String("ec2.amazonaws.com"), nil),
		Description: jsii.String("Role for EC2 instances running benchmarks"),
	})

	// Add SSM managed policy
	ec2InstanceRole.AddManagedPolicy(awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonSSMManagedInstanceCore")))
	// Add CloudWatch managed policy
	ec2InstanceRole.AddManagedPolicy(awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("CloudWatchAgentServerPolicy")))

	// grant read permissions to the binary s3 asset
	binaryS3Asset.GrantRead(ec2InstanceRole)

	// Grant read/write permissions to the specific results bucket
	resultsBucket.GrantReadWrite(ec2InstanceRole, "*")

	// Add the ExportResults Lambda function
	exportResultsLambda := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("ExportResultsLambda"), &awscdklambdagoalpha.GoFunctionProps{
		Entry:   jsii.String("./stacks/benchmark/lambdas/exportresults"),
		Runtime: awslambda.Runtime_PROVIDED_AL2023(),
		Timeout: awscdk.Duration_Minutes(jsii.Number(5)),
		Bundling: &awscdklambdagoalpha.BundlingOptions{
			GoBuildFlags: &[]*string{
				jsii.String("-ldflags \"-s -w\""),
			},
			Platform: jsii.String("linux/amd64"),
		},
	})

	// grant the lambda function permission to write to the results bucket
	resultsBucket.GrantReadWrite(exportResultsLambda, "*")

	// Use default key pair
	keyPairName := config.KeyPairName(scope)
	if len(keyPairName) == 0 {
		panic("KeyPairName is empty")
	}

	keyPair := awsec2.KeyPair_FromKeyPairName(
		stack,
		jsii.String("BenchmarkKeyPair"),
		jsii.String(keyPairName),
	)

	// Create EC2 launch templates for each instance type
	launchTemplatesMap := make(map[awsec2.InstanceType]CreateLaunchTemplateOutput)
	for _, instanceType := range testedInstances {
		launchTemplatesMap[instanceType] = createLaunchTemplate(
			stack,
			CreateLaunchTemplateInput{
				ID:            fmt.Sprintf("benchmark-%s", *instanceType.ToString()),
				InstanceType:  instanceType,
				BinaryS3Asset: binaryS3Asset,
				SecurityGroup: securityGroup,
				IAMRole:       ec2InstanceRole,
				KeyPair:       keyPair,
			},
		)
	}

	// Create the main state machine to orchestrate the benchmark process
	stateMachine := createStateMachine(stack, CreateStateMachineInput{
		LaunchTemplatesMap:  launchTemplatesMap,
		BinaryS3Asset:       binaryS3Asset,
		ResultsBucket:       resultsBucket,
		ExportResultsLambda: exportResultsLambda,
	})

	// Add permissions to the state machine's execution role
	stateMachine.Role().AddToPrincipalPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect: awsiam.Effect_ALLOW,
		Actions: jsii.Strings(
			"ec2:RunInstances",
			"ec2:TerminateInstances",
			"ec2:DescribeInstances",
			"ec2:DescribeInstanceStatus",
			"ec2:CreateTags",
			"ec2:DescribeKeyPairs",
			"ec2:DescribeLaunchTemplates",
			"ec2:DescribeLaunchTemplateVersions",
			"iam:PassRole",
		),
		Resources: jsii.Strings("*"),
	}))

	// Add a separate statement for PassRole against the ec2 instance role
	stateMachine.Role().AddToPrincipalPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect:    awsiam.Effect_ALLOW,
		Actions:   jsii.Strings("iam:PassRole"),
		Resources: jsii.Strings(*ec2InstanceRole.RoleArn()),
	}))

	// Add permissions for SSM
	stateMachine.Role().AddToPrincipalPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect: awsiam.Effect_ALLOW,
		Actions: jsii.Strings(
			"ssm:SendCommand",
			"ssm:GetCommandInvocation",
		),
		Resources: jsii.Strings("*"),
	}))

	return stack
}
