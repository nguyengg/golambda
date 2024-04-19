package getenv

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

// SecretsOpts contains customisable settings when retrieving a variable from AWS Secrets Manager.
type SecretsOpts struct {
	SecretId     string
	VersionId    string
	VersionStage string
	Client       http.Client
}

// Secrets creates Getter that reads secrets from the AWS Parameter and Secrets Lambda extension.
//
// If you need to customize the request with version, label, and/or with decryption, pass in a function to modify those values.
//
// See https://docs.aws.amazon.com/secretsmanager/latest/userguide/retrieving-secrets_lambda.html.
func Secrets(secretId string, opts ...func(*SecretsOpts)) Variable {
	g, err := NewSecretsGetter(secretId, opts...)
	if err != nil {
		return errVar{err: err}
	}

	return Getter(func(ctx context.Context) (string, error) {
		output, err := g.Get(ctx)
		if err != nil {
			return "", err
		}

		return aws.ToString(output.SecretString), nil
	})
}

type SecretsGetter struct {
	client http.Client
	req    *http.Request
}

// NewSecretsGetter returns an instance of SecretsGetter that can be used to get the raw secretsmanager.GetSecretValueOutput.
func NewSecretsGetter(secretId string, opts ...func(secretsOpts *SecretsOpts)) (*SecretsGetter, error) {
	port := os.Getenv("PARAMETERS_SECRETS_EXTENSION_HTTP_PORT")
	if port == "" {
		return nil, fmt.Errorf("no PARAMETERS_SECRETS_EXTENSION_HTTP_PORT")
	}
	if _, err := strconv.ParseInt(port, 10, 64); err != nil {
		return nil, fmt.Errorf("PARAMETERS_SECRETS_EXTENSION_HTTP_PORT is not an integer: %w", err)
	}

	token := os.Getenv("AWS_SESSION_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("no AWS_SESSION_TOKEN")
	}

	params := SecretsOpts{
		SecretId: secretId,
		Client:   http.Client{},
	}
	for _, opt := range opts {
		opt(&params)
	}

	req, err := http.NewRequest("GET", "http://localhost:"+port+"/secretsmanager/get", nil)
	if err != nil {
		return nil, fmt.Errorf("create GET secrets request error: %w", err)
	}

	req.Header.Add("X-Aws-Parameters-Secrets-Token", token)

	q := url.Values{}
	q.Add("secretId", params.SecretId)
	if params.VersionId != "" {
		q.Add("versionId", params.VersionId)
	}
	if params.VersionStage != "" {
		q.Add("versionStage", params.VersionStage)
	}
	req.URL.RawQuery = q.Encode()

	return &SecretsGetter{
		client: params.Client,
		req:    req,
	}, nil
}

// Get executes the GET request the AWS Parameter and Secrets Lambda extension.
func (g *SecretsGetter) Get(ctx context.Context) (*secretsmanager.GetSecretValueOutput, error) {
	res, err := g.client.Do(g.req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("do GET secrets error: %w", err)
	}

	output := &secretsmanager.GetSecretValueOutput{}
	err = json.NewDecoder(res.Body).Decode(output)
	_ = res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("decode GET secrets response error: %w", err)
	}

	return output, nil
}
