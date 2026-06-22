package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2integrations"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscognito"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

const (
	region            = "ap-south-1"
	cognitoUserPoolID = "ap-south-1_7HEDrxYBc"

	// lambdaAsset is the directory the API service builds into (`make build`
	// in services/api produces ./build/bootstrap). Path is relative to this file.
	lambdaAsset = "../services/api/build"
)

func newApiStack(scope constructs.Construct, id string, props *awscdk.StackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, props)

	// App clients for the two native mobile apps, both on the EXISTING user pool.
	// No secret (public native clients) and SRP-only auth, matching
	// amazon-cognito-identity-js's authenticateUser flow. The API's Cognito
	// middleware accepts any valid access token from this pool regardless of
	// which client issued it (COGNITO_CLIENT_ID is intentionally left unset
	// below), so adding more clients here never requires a backend change.
	pool := awscognito.UserPool_FromUserPoolId(stack, jsii.String("UserPool"), jsii.String(cognitoUserPoolID))
	driverClient := pool.AddClient(jsii.String("DriverAppClient"), &awscognito.UserPoolClientOptions{
		UserPoolClientName: jsii.String("raftaar-driver-app"),
		GenerateSecret:     jsii.Bool(false),
		AuthFlows:          &awscognito.AuthFlow{UserSrp: jsii.Bool(true)},
	})
	userClient := pool.AddClient(jsii.String("UserAppClient"), &awscognito.UserPoolClientOptions{
		UserPoolClientName: jsii.String("raftaar-user-app"),
		GenerateSecret:     jsii.Bool(false),
		AuthFlows:          &awscognito.AuthFlow{UserSrp: jsii.Bool(true)},
	})

	// Single-table DynamoDB store (on-demand) with a GSI for secondary lookups.
	table := awsdynamodb.NewTableV2(stack, jsii.String("Table"), &awsdynamodb.TablePropsV2{
		TableName:    jsii.String("raftaar-users"),
		PartitionKey: &awsdynamodb.Attribute{Name: jsii.String("PK"), Type: awsdynamodb.AttributeType_STRING},
		SortKey:      &awsdynamodb.Attribute{Name: jsii.String("SK"), Type: awsdynamodb.AttributeType_STRING},
		Billing:      awsdynamodb.Billing_OnDemand(nil),
		// DESTROY so `cdk destroy` cleans up; switch to RETAIN before going to prod.
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		GlobalSecondaryIndexes: &[]*awsdynamodb.GlobalSecondaryIndexPropsV2{
			{
				IndexName:    jsii.String("GSI1"),
				PartitionKey: &awsdynamodb.Attribute{Name: jsii.String("GSI1PK"), Type: awsdynamodb.AttributeType_STRING},
				SortKey:      &awsdynamodb.Attribute{Name: jsii.String("GSI1SK"), Type: awsdynamodb.AttributeType_STRING},
			},
		},
	})

	// API Lambda: custom runtime (provided.al2023) running the Go `bootstrap` binary.
	fn := awslambda.NewFunction(stack, jsii.String("ApiFn"), &awslambda.FunctionProps{
		FunctionName: jsii.String("raftaar-api"),
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		Architecture: awslambda.Architecture_ARM_64(),
		Handler:      jsii.String("bootstrap"),
		Code:         awslambda.Code_FromAsset(jsii.String(lambdaAsset), nil),
		MemorySize:   jsii.Number(256),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(15)),
		LogGroup: awslogs.NewLogGroup(stack, jsii.String("ApiFnLogs"), &awslogs.LogGroupProps{
			Retention:     awslogs.RetentionDays_ONE_WEEK,
			RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		}),
		Environment: &map[string]*string{
			// AWS_REGION is provided by the runtime; pass it explicitly for local parity.
			"COGNITO_REGION":       jsii.String(region),
			"COGNITO_USER_POOL_ID": jsii.String(cognitoUserPoolID),
			"DYNAMODB_TABLE":       table.TableName(),
		},
	})

	// Grant the Lambda read/write access to the table (and its indexes).
	table.Grants().ReadWriteData(fn)

	// HTTP API (payload v2.0) with the Lambda as the catch-all default integration.
	integration := awsapigatewayv2integrations.NewHttpLambdaIntegration(jsii.String("ApiIntegration"), fn, nil)
	api := awsapigatewayv2.NewHttpApi(stack, jsii.String("HttpApi"), &awsapigatewayv2.HttpApiProps{
		ApiName:            jsii.String("raftaar-api"),
		DefaultIntegration: integration,
	})

	awscdk.NewCfnOutput(stack, jsii.String("ApiUrl"), &awscdk.CfnOutputProps{
		Value: api.ApiEndpoint(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("DriverAppClientId"), &awscdk.CfnOutputProps{
		Value: driverClient.UserPoolClientId(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("UserAppClientId"), &awscdk.CfnOutputProps{
		Value: userClient.UserPoolClientId(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("TableName"), &awscdk.CfnOutputProps{
		Value: table.TableName(),
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	newApiStack(app, "RaftaarApiStack", &awscdk.StackProps{
		Env: &awscdk.Environment{
			Region: jsii.String(region),
		},
	})

	app.Synth(nil)
}
