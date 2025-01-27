package opaquetoken

import (
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"reflect"
	"testing"
)

func TestConverter_Encode(t *testing.T) {
	type fields struct {
		Transformer Transformer
	}
	type args struct {
		key map[string]dynamodbtypes.AttributeValue
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:   "S hash, B sort",
			fields: fields{nil},
			args: args{key: map[string]dynamodbtypes.AttributeValue{
				"id":    &dynamodbtypes.AttributeValueMemberS{Value: "hash"},
				"range": &dynamodbtypes.AttributeValueMemberB{Value: []byte("hello, world!")},
			}},
			want: `{"id":{"S":"hash"},"range":{"B":"aGVsbG8sIHdvcmxkIQ"}}`,
		},
		{
			name:   "B hash, N sort",
			fields: fields{nil},
			args: args{key: map[string]dynamodbtypes.AttributeValue{
				"id":      &dynamodbtypes.AttributeValueMemberB{Value: []byte("hello, world!")},
				"version": &dynamodbtypes.AttributeValueMemberN{Value: "42"},
			}},
			want: `{"id":{"B":"aGVsbG8sIHdvcmxkIQ"},"version":{"N":"42"}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Converter{
				Transformer: tt.fields.Transformer,
			}
			got, err := c.Encode(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Encode() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConverter_Decode(t *testing.T) {
	type fields struct {
		Transformer Transformer
	}
	type args struct {
		token string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantKey map[string]dynamodbtypes.AttributeValue
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:   "S hash, B sort",
			fields: fields{nil},
			args:   args{`{"id":{"S":"hash"},"range":{"B":"aGVsbG8sIHdvcmxkIQ"}}`},
			wantKey: map[string]dynamodbtypes.AttributeValue{
				"id":    &dynamodbtypes.AttributeValueMemberS{Value: "hash"},
				"range": &dynamodbtypes.AttributeValueMemberB{Value: []byte("hello, world!")},
			},
		},
		{
			name:   "B hash, N sort",
			fields: fields{nil},
			args:   args{`{"id":{"B":"aGVsbG8sIHdvcmxkIQ"},"version":{"N":"42"}}`},
			wantKey: map[string]dynamodbtypes.AttributeValue{
				"id":      &dynamodbtypes.AttributeValueMemberB{Value: []byte("hello, world!")},
				"version": &dynamodbtypes.AttributeValueMemberN{Value: "42"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Converter{
				Transformer: tt.fields.Transformer,
			}
			gotKey, err := c.Decode(tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotKey, tt.wantKey) {
				t.Errorf("Decode() gotKey = %v, want %v", gotKey, tt.wantKey)
			}
		})
	}
}
