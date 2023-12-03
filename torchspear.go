package main

import (
	"os"
	"path/filepath"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfrontorigins"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53targets"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type TorchspearRedirectStackProps struct {
	awscdk.StackProps
}

func NewTorchspearRedirectStack(scope constructs.Construct, id string, props *TorchspearRedirectStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	zone := awsroute53.HostedZone_FromLookup(stack, jsii.String("TorchspearComZone"), &awsroute53.HostedZoneProviderProps{
		DomainName: jsii.String("torchspear.com"),
	})

	cert := awscertificatemanager.NewCertificate(stack, jsii.String("TorchspearComCert"), &awscertificatemanager.CertificateProps{
		DomainName: jsii.String("torchspear.com"),
		SubjectAlternativeNames: &[]*string{
			jsii.String("www.torchspear.com"),
		},
		CertificateName: jsii.String("torchspear.com CDN Certificate"),
		Validation:      awscertificatemanager.CertificateValidation_FromDns(zone),
	})

	workDir, err := os.Getwd()
	if err != nil {
		workDir = "."
	}

	cfFunc := awscloudfront.NewFunction(stack, jsii.String("RedirectFunction"), &awscloudfront.FunctionProps{
		Code: awscloudfront.FunctionCode_FromFile(&awscloudfront.FileCodeOptions{
			FilePath: jsii.String(filepath.Join(workDir, "func", "redirect.js")),
		}),
	})

	cdn := awscloudfront.NewDistribution(stack, jsii.String("CloudfrontDistribution"), &awscloudfront.DistributionProps{
		DefaultBehavior: &awscloudfront.BehaviorOptions{
			Origin: awscloudfrontorigins.NewHttpOrigin(jsii.String("null.invalid"), &awscloudfrontorigins.HttpOriginProps{}),
			FunctionAssociations: &[]*awscloudfront.FunctionAssociation{
				{
					Function:  cfFunc,
					EventType: awscloudfront.FunctionEventType_VIEWER_REQUEST,
				},
			},
		},
		DomainNames: &[]*string{
			jsii.String("torchspear.com"),
			jsii.String("www.torchspear.com"),
		},
		Certificate: cert,
	})

	awsroute53.NewAaaaRecord(stack, jsii.String("AaaaRecord"), &awsroute53.AaaaRecordProps{
		Zone:   zone,
		Target: awsroute53.RecordTarget_FromAlias(awsroute53targets.NewCloudFrontTarget(cdn)),
	})

	awsroute53.NewARecord(stack, jsii.String("ARecord"), &awsroute53.ARecordProps{
		Zone:   zone,
		Target: awsroute53.RecordTarget_FromAlias(awsroute53targets.NewCloudFrontTarget(cdn)),
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewTorchspearRedirectStack(app, "TorchspearRedirectStack", &TorchspearRedirectStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}
