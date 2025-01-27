package opaquetoken

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Tokenizer converts from DynamoDB's last evaluated key to pagination token and vice versa for query and scan operations.
//
// The default value is ready for use without any encryption. Prefer NewWithAES to conform to opaque token principle.
//
// Per specifications, only three data types (S, N, or B) can be partition key or sort key. The pagination token will
// be the DynamoDB JSON blob of the evaluated key, which should have no more than 2 entries.
type Tokenizer struct {
	// Transformer can be used to encrypt/decrypt the tokens to conform to opaque token principle.
	Transformer Transformer
}

// Encode converts the given last evaluated key to the pagination token.
func (t Tokenizer) Encode(key map[string]dynamodbtypes.AttributeValue) (string, error) {
	switch n := len(key); n {
	case 1, 2:
	default:
		return "", fmt.Errorf("invalid number of attributes in key: expected 1 or 2, got (%d)", n)
	}

	item := make(map[string]map[string]string)
	for k, v := range key {
		item[k] = make(map[string]string)

		avS, ok := v.(*dynamodbtypes.AttributeValueMemberS)
		if ok {
			item[k]["S"] = avS.Value
			continue
		}

		avN, ok := v.(*dynamodbtypes.AttributeValueMemberN)
		if ok {
			item[k]["N"] = avN.Value
			continue
		}

		avB, ok := v.(*dynamodbtypes.AttributeValueMemberB)
		if ok {
			item[k]["B"] = base64.RawStdEncoding.EncodeToString(avB.Value)
			continue
		}

		return "", fmt.Errorf("key named %s has unknown type %T", k, v)
	}

	token, err := json.Marshal(item)
	if err != nil {
		return "", fmt.Errorf("marshal token as JSON error: %w", err)
	}

	if t.Transformer != nil {
		return t.Transformer.Encode(string(token))
	}

	return string(token), nil
}

// Decode converts the given pagination token to exclusive start key.
func (t Tokenizer) Decode(token string) (key map[string]dynamodbtypes.AttributeValue, err error) {
	if t.Transformer != nil {
		if token, err = t.Transformer.Decode(token); err != nil {
			return
		}
	}

	item := make(map[string]map[string]string)
	if err = json.Unmarshal([]byte(token), &item); err != nil {
		return nil, fmt.Errorf("unmarshal token as JSON error: %w", err)
	}

	switch n := len(item); n {
	case 1, 2:
	default:
		return nil, fmt.Errorf("invalid number of attributes in key: expected 1 or 2, got (%d)", n)
	}

	key = make(map[string]dynamodbtypes.AttributeValue)
	for k, v := range item {
		if n := len(v); n != 1 {
			return nil, fmt.Errorf("invalid number of attributes in key named %s: expected 1 or 2, got (%d)", k, n)
		}

		avS, ok := v["S"]
		if ok {
			key[k] = &dynamodbtypes.AttributeValueMemberS{Value: avS}
			continue
		}

		avN, ok := v["N"]
		if ok {
			key[k] = &dynamodbtypes.AttributeValueMemberN{Value: avN}
			continue
		}

		avB, ok := v["B"]
		if ok {
			data, err := base64.RawStdEncoding.DecodeString(avB)
			if err != nil {
				return nil, fmt.Errorf("decode attribute named %s as B error: %w", k, err)
			}

			key[k] = &dynamodbtypes.AttributeValueMemberB{Value: data}
			continue
		}
	}

	return key, nil
}
