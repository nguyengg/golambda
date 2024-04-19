package getenv

import (
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
}

// Parameter creates Getter that reads parameters from the AWS Parameter and Secrets Lambda extension.
//
// If you need to customize the request with version, label, and/or with decryption, pass in a function to modify those values.
//
// See https://docs.aws.amazon.com/systems-manager/latest/userguide/ps-integration-lambda-extensions.html#ps-integration-lambda-extensions-sample-commands.
func Parameter(name string, opts ...func(*ParameterOpts)) Variable {
	var err error
	port := os.Getenv("PARAMETERS_SECRETS_EXTENSION_HTTP_PORT")
	if port == "" {
		return errVar{err: fmt.Errorf("no PARAMETERS_SECRETS_EXTENSION_HTTP_PORT")}
	}
	if _, err := strconv.ParseInt(port, 10, 64); err != nil {
		return errVar{err: fmt.Errorf("PARAMETERS_SECRETS_EXTENSION_HTTP_PORT is not an integer: %w", err)}
	}

	token := os.Getenv("AWS_SESSION_TOKEN")
	if token == "" {
		return errVar{err: fmt.Errorf("no AWS_SESSION_TOKEN")}
	}

	params := ParameterOpts{Name: name}
	for _, opt := range opts {
		opt(&params)
	}

	client := http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:"+port+"/systemsmanager/parameters/get", nil)
	if err != nil {
		return errVar{err: fmt.Errorf("create HTTP request error: %w", err)}
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

	return getter(func() (string, error) {
		res, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("GET parameter store error: %w", err)
		}

		output := ssm.GetParameterOutput{}
		err = json.NewDecoder(res.Body).Decode(&output)
		_ = res.Body.Close()
		if err != nil {
			return "", fmt.Errorf("decode GET parameter store error: %w", err)
		}

		return aws.ToString(output.Parameter.Value), nil
	})
}
