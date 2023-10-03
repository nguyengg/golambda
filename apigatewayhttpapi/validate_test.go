package apigatewayhttpapi

import (
	. "context"
	. "github.com/aws/aws-lambda-go/events"
	"testing"
)

func TestValidate_Success(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		take    in
		give    out
		wantErr bool
	}{
		// https://docs.aws.amazon.com/lambda/latest/dg/golang-handler.html#golang-handler-signatures
		{
			name:    "func ()",
			args:    args{v: func() {}},
			take:    receiveNothing,
			give:    returnNothing,
			wantErr: false,
		},
		{
			name:    "func () (err)",
			args:    args{v: func() error { return nil }},
			take:    receiveNothing,
			give:    returnError,
			wantErr: false,
		},
		{
			name:    "func (TIn) error",
			args:    args{v: func(APIGatewayV2HTTPRequest) error { return nil }},
			take:    receiveHttpV2Request,
			give:    returnError,
			wantErr: false,
		},
		{
			name:    "func () (TOut, err)",
			args:    args{v: func() (APIGatewayV2HTTPResponse, error) { return APIGatewayV2HTTPResponse{}, nil }},
			take:    receiveNothing,
			give:    returnHttpV2Response | returnError,
			wantErr: false,
		},
		{
			name:    "func (Context) (err)",
			args:    args{v: func(Context) error { return nil }},
			take:    receiveContext,
			give:    returnError,
			wantErr: false,
		},
		{
			name: "func (Context, TIn) (TOut, err)",
			args: args{v: func(Context, APIGatewayV2HTTPRequest) (APIGatewayV2HTTPResponse, error) {
				return APIGatewayV2HTTPResponse{}, nil
			}},
			take:    receiveContext | receiveHttpV2Request,
			give:    returnHttpV2Response | returnError,
			wantErr: false,
		},
		{
			name:    "func (Context) (TOut, err)",
			args:    args{v: func(Context) (APIGatewayV2HTTPResponse, error) { return APIGatewayV2HTTPResponse{}, nil }},
			take:    receiveContext,
			give:    returnHttpV2Response | returnError,
			wantErr: false,
		},
		{
			name:    "func (Context, TIn) (err)",
			args:    args{v: func(Context, APIGatewayV2HTTPRequest) error { return nil }},
			take:    receiveContext | receiveHttpV2Request,
			give:    returnError,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			take, give, err := validate(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if take != tt.take {
				t.Errorf("validate() take = %v, want %v", take, tt.take)
			}
			if give != tt.give {
				t.Errorf("validate() give = %v, want %v", give, tt.give)
			}
		})
	}
}

func TestValidate_error(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		err  string
	}{
		{
			name: "nil handler",
			args: args{v: nil},
		},
		{
			name: "handler must be a function instead of type int64",
			args: args{v: int64(123)},
		},
		{
			name: "handler must be a function instead of type struct",
			args: args{v: struct{ hello string }{hello: "world"}},
		},
		{
			name: "the sole argument must be either context.Context or events.APIGatewayV2HTTPRequest instead of string",
			args: args{v: func(string) {}},
		},
		{
			name: "the 1st argument must be context.Context instead of string",
			args: args{v: func(string, string) {}},
		},
		{
			name: "the 2nd argument must be events.APIGatewayV2HTTPRequest instead of string",
			args: args{v: func(Context, string) {}},
		},
		{
			name: "can only receive up to 2 arguments instead of 3",
			args: args{v: func(string, string, string) {}},
		},
		{
			name: "the sole returned value must be error instead of string",
			args: args{v: func() string { return "" }},
		},
		{
			name: "the 2nd returned value must be error instead of string",
			args: args{v: func() (string, string) { return "", "" }},
		},
		{
			name: "can only return up to 2 values instead of 3",
			args: args{v: func() (string, string, string) { return "", "", "" }},
		},
		{
			name: "handler must return at least one value if it also takes at least one argument",
			args: args{v: func(Context) {}},
		},
		{
			name: "if the sole argument is events.APIGatewayV2HTTPRequest then handler must return exactly one value that implements error type",
			args: args{v: func(APIGatewayV2HTTPRequest) (APIGatewayV2HTTPResponse, error) {
				return APIGatewayV2HTTPResponse{}, nil
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := validate(tt.args.v)
			if err == nil {
				t.Errorf("validate() want error")
				return
			} else if err.Error() != tt.name {
				t.Errorf("validate() error.Error() actual=%q, expected=%q", err.Error(), tt.name)
				return
			}
		})
	}
}
