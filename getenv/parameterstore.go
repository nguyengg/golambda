package getenv

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

// ParameterOpts contains customisable settings when retrieving a variable from AWS Parameter Store.
type ParameterOpts struct {
	Name           string
	Version        string
	Label          string
	WithDecryption bool
	Client         http.Client
}

// Parameter creates Getter that reads parameters from the AWS Parameter and Secrets Lambda extension.
//
// If you need to customize the request with version, label, and/or with decryption, pass in a function to modify those values.
//
// See https://docs.aws.amazon.com/systems-manager/latest/userguide/ps-integration-lambda-extensions.html#ps-integration-lambda-extensions-sample-commands.
func Parameter(name string, opts ...func(*ParameterOpts)) Variable[string] {
	g, err := NewParameterGetter(name, opts...)
	if err != nil {
		return errVar[string]{err: err}
	}

	return Getter(func(ctx context.Context) (string, error) {
		output, err := g.Get(ctx)
		if err != nil {
			return "", err
		}

		return aws.ToString(output.Parameter.Value), nil
	})
}

type ParameterGetter struct {
	client http.Client
	req    *http.Request
}

// NewParameterGetter returns an instance of ParameterGetter that can be used to get the raw ssm.GetParameterOutput.
func NewParameterGetter(name string, opts ...func(*ParameterOpts)) (*ParameterGetter, error) {
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

	params := ParameterOpts{
		Name:   name,
		Client: http.Client{},
	}
	for _, opt := range opts {
		opt(&params)
	}

	req, err := http.NewRequest("GET", "http://localhost:"+port+"/systemsmanager/parameters/get", nil)
	if err != nil {
		return nil, fmt.Errorf("create GET parameter store request error: %w", err)
	}

	req.Header.Add("X-Aws-Parameters-Secrets-Token", token)

	q := url.Values{}
	q.Add("name", params.Name)
	if params.Version != "" {
		q.Add("version", params.Version)
	}
	if params.Label != "" {
		q.Add("label", params.Label)
	}
	if params.WithDecryption {
		q.Add("withDecryption", "true")
	}
	req.URL.RawQuery = q.Encode()

	return &ParameterGetter{
		client: params.Client,
		req:    req,
	}, nil
}

// Get executes the GET request the AWS Parameter and Secrets Lambda extension.
func (g *ParameterGetter) Get(ctx context.Context) (*ssm.GetParameterOutput, error) {
	res, err := g.client.Do(g.req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("do GET parameter store error: %w", err)
	}

	output := &ssm.GetParameterOutput{}
	err = json.NewDecoder(res.Body).Decode(output)
	_ = res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("decode GET parameter store response error: %w", err)
	}

	return output, nil
}
