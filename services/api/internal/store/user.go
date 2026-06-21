package store

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// User is the domain model for a user profile, keyed by the Cognito subject.
type User struct {
	Sub       string    `json:"sub"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Users is the repository for User records in the single-table design.
//
// Key layout:
//
//	PK = "USER#<sub>"   SK = "PROFILE"                  -> profile by subject
//	GSI1PK = "PHONE#<phone>"  GSI1SK = "USER#<sub>"     -> lookup by phone (GSI1)
type Users struct {
	db    *dynamodb.Client
	table string
}

// NewUsers returns a Users repository bound to the given table.
func NewUsers(db *dynamodb.Client, table string) *Users {
	return &Users{db: db, table: table}
}

// userItem is the on-the-wire DynamoDB representation including the table keys.
type userItem struct {
	PK        string `dynamodbav:"PK"`
	SK        string `dynamodbav:"SK"`
	GSI1PK    string `dynamodbav:"GSI1PK,omitempty"`
	GSI1SK    string `dynamodbav:"GSI1SK,omitempty"`
	Type      string `dynamodbav:"Type"`
	Sub       string `dynamodbav:"Sub"`
	Name      string `dynamodbav:"Name"`
	Phone     string `dynamodbav:"Phone"`
	CreatedAt string `dynamodbav:"CreatedAt"`
	UpdatedAt string `dynamodbav:"UpdatedAt"`
}

func userPK(sub string) string { return "USER#" + sub }

const (
	userSK       = "PROFILE"
	userType     = "User"
	phoneIndex   = "GSI1"
	phoneGSI1Key = "PHONE#"
)

// Get returns the profile for a Cognito subject, or ErrNotFound.
func (r *Users) Get(ctx context.Context, sub string) (User, error) {
	out, err := r.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: userPK(sub)},
			"SK": &types.AttributeValueMemberS{Value: userSK},
		},
	})
	if err != nil {
		return User{}, fmt.Errorf("get user: %w", err)
	}
	if out.Item == nil {
		return User{}, ErrNotFound
	}
	return toUser(out.Item)
}

// Put upserts a user profile. CreatedAt is set on first write; UpdatedAt is
// always refreshed. The returned User reflects the persisted timestamps.
func (r *Users) Put(ctx context.Context, u User) (User, error) {
	now := time.Now().UTC()
	if u.CreatedAt.IsZero() {
		u.CreatedAt = now
	}
	u.UpdatedAt = now

	item := userItem{
		PK:        userPK(u.Sub),
		SK:        userSK,
		Type:      userType,
		Sub:       u.Sub,
		Name:      u.Name,
		Phone:     u.Phone,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
	}
	if u.Phone != "" {
		item.GSI1PK = phoneGSI1Key + u.Phone
		item.GSI1SK = userPK(u.Sub)
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return User{}, fmt.Errorf("marshal user: %w", err)
	}
	if _, err := r.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.table),
		Item:      av,
	}); err != nil {
		return User{}, fmt.Errorf("put user: %w", err)
	}
	return u, nil
}

// GetByPhone looks up a user via the GSI1 phone index, or ErrNotFound.
func (r *Users) GetByPhone(ctx context.Context, phone string) (User, error) {
	out, err := r.db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.table),
		IndexName:              aws.String(phoneIndex),
		KeyConditionExpression: aws.String("GSI1PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: phoneGSI1Key + phone},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return User{}, fmt.Errorf("query user by phone: %w", err)
	}
	if len(out.Items) == 0 {
		return User{}, ErrNotFound
	}
	return toUser(out.Items[0])
}

func toUser(item map[string]types.AttributeValue) (User, error) {
	var it userItem
	if err := attributevalue.UnmarshalMap(item, &it); err != nil {
		return User{}, fmt.Errorf("unmarshal user: %w", err)
	}
	created, _ := time.Parse(time.RFC3339, it.CreatedAt)
	updated, _ := time.Parse(time.RFC3339, it.UpdatedAt)
	return User{
		Sub:       it.Sub,
		Name:      it.Name,
		Phone:     it.Phone,
		CreatedAt: created,
		UpdatedAt: updated,
	}, nil
}
