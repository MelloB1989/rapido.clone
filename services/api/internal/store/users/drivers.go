package users

import (
	"apiservice/internal/config"
	"apiservice/internal/errors"
	"apiservice/internal/shared"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Rider struct {
	Sub       string    `json:"sub"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Riders struct {
	deps  *shared.Deps
	table string
}

func NewRiders(deps *shared.Deps) *Riders {
	return &Riders{deps: deps, table: config.Load().UsersTableName}
}

type riderItem struct {
	PK                string `dynamodbav:"PK"`
	SK                string `dynamodbav:"SK"`
	GSI1PK            string `dynamodbav:"GSI1PK,omitempty"`
	GSI1SK            string `dynamodbav:"GSI1SK,omitempty"`
	Type              string `dynamodbav:"Type"`
	Sub               string `dynamodbav:"Sub"`
	Name              string `dynamodbav:"Name"`
	DriversLicenseUrl string `dynamodbav:"DriversLicenseUrl"`
	Phone             string `dynamodbav:"Phone"`
	CreatedAt         string `dynamodbav:"CreatedAt"`
	UpdatedAt         string `dynamodbav:"UpdatedAt"`
}

func riderPK(sub string) string { return "RIDER#" + sub }

const (
	riderSK   = "PROFILE"
	riderType = "Rider"
)

// Get returns the profile for a Cognito subject, or ErrNotFound.
func (r *Riders) Get(ctx context.Context, sub string) (*Rider, errors.StoreError) {
	if r.deps.Check() {
		return nil, errors.ClientNotInitialized
	}
	out, err := r.deps.DDB.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: userPK(sub)},
			"SK": &types.AttributeValueMemberS{Value: userSK},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if out.Item == nil {
		return nil, errors.ErrNotFound
	}
	return toRider(out.Item)
}

// Put upserts a user profile. CreatedAt is set on first write; UpdatedAt is
// always refreshed. The returned User reflects the persisted timestamps.
func (r *Riders) Put(ctx context.Context, u Rider) (*Rider, errors.StoreError) {
	now := time.Now().UTC()
	if u.CreatedAt.IsZero() {
		u.CreatedAt = now
	}
	u.UpdatedAt = now

	item := riderItem{
		PK:                userPK(u.Sub),
		SK:                userSK,
		Type:              userType,
		Sub:               u.Sub,
		Name:              u.Name,
		Phone:             u.Phone,
		DriversLicenseUrl: "",
		CreatedAt:         u.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         u.UpdatedAt.Format(time.RFC3339),
	}
	if u.Phone != "" {
		item.GSI1PK = phoneGSI1Key + u.Phone
		item.GSI1SK = userPK(u.Sub)
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return nil, fmt.Errorf("marshal user: %w", err)
	}
	if _, err := r.deps.DDB.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.table),
		Item:      av,
	}); err != nil {
		return nil, fmt.Errorf("put user: %w", err)
	}
	return &u, nil
}

// GetByPhone looks up a user via the GSI1 phone index, or ErrNotFound.
func (r *Riders) GetByPhone(ctx context.Context, phone string) (*Rider, errors.StoreError) {
	out, err := r.deps.DDB.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.table),
		IndexName:              aws.String(phoneIndex),
		KeyConditionExpression: aws.String("GSI1PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: phoneGSI1Key + phone},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, fmt.Errorf("query user by phone: %w", err)
	}
	if len(out.Items) == 0 {
		return nil, errors.ErrNotFound
	}
	return toRider(out.Items[0])
}

func toRider(item map[string]types.AttributeValue) (*Rider, error) {
	var it userItem
	if err := attributevalue.UnmarshalMap(item, &it); err != nil {
		return nil, fmt.Errorf("unmarshal user: %w", err)
	}
	created, _ := time.Parse(time.RFC3339, it.CreatedAt)
	updated, _ := time.Parse(time.RFC3339, it.UpdatedAt)
	return &Rider{
		Sub:       it.Sub,
		Name:      it.Name,
		Phone:     it.Phone,
		CreatedAt: created,
		UpdatedAt: updated,
	}, nil
}
