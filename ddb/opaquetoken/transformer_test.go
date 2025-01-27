package opaquetoken

import (
	"crypto/aes"
	"crypto/cipher"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nguyengg/golambda/must"
	"reflect"
	"testing"
)

func Test_aesTransformer_EncodeDecode(t *testing.T) {
	key := []byte("onvIzKsW6Ec2Q5VqS49zrNlmvrvibh8e")

	type fields struct {
		c cipher.Block
	}
	type args struct {
		s string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "test",
			fields: fields{c: must.Must(aes.NewCipher(key))},
			args:   args{"hello, world!"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := aesTransformer{
				c: tt.fields.c,
			}

			got, err := a.Encode(tt.args.s)
			if err != nil {
				t.Errorf("Encode() error = %v", err)
				return
			}

			got, err = a.Decode(got)
			if err != nil {
				t.Errorf("Decode() error = %v", err)
				return
			}

			if got != tt.args.s {
				t.Errorf("Decode() got = %v, want %v", got, tt.args.s)
			}
		})
	}
}

func TestTokenizer_EncodeDecodeWithAES(t *testing.T) {
	key := []byte("onvIzKsW6Ec2Q5VqS49zrNlmvrvibh8e")

	type args struct {
		key map[string]dynamodbtypes.AttributeValue
	}

	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			name: "S hash, B sort",
			args: args{key: map[string]dynamodbtypes.AttributeValue{
				"id":    &dynamodbtypes.AttributeValueMemberS{Value: "hash"},
				"range": &dynamodbtypes.AttributeValueMemberB{Value: []byte("hello, world!")},
			}},
		},
		{
			name: "B hash, N sort",
			args: args{key: map[string]dynamodbtypes.AttributeValue{
				"id":      &dynamodbtypes.AttributeValueMemberB{Value: []byte("hello, world!")},
				"version": &dynamodbtypes.AttributeValueMemberN{Value: "42"},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenizer, err := NewWithAES(key)
			if err != nil {
				t.Errorf("NewWithAES() error = %v", err)
				return
			}

			token, err := tokenizer.Encode(tt.args.key)
			if err != nil {
				t.Errorf("Encode() error = %v", err)
				return
			}

			got, err := tokenizer.Decode(token)
			if err != nil {
				t.Errorf("Decode() error = %v", err)
				return
			}

			if !reflect.DeepEqual(got, tt.args.key) {
				t.Errorf("Decode() got = %v, want %v", got, tt.args.key)
			}
		})
	}
}
