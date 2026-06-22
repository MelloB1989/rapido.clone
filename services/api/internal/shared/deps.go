package shared

import "github.com/aws/aws-sdk-go-v2/service/dynamodb"

type Deps struct {
	DDB *dynamodb.Client
}

func (d *Deps) Check() bool {
	return !(d.DDB == nil)
}
