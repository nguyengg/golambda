package getenv

import (
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
}

// Secrets creates Getter that reads secrets from the AWS Parameter and Secrets Lambda extension.
//
// If you need to customize the request with version, label, and/or with decryption, pass in a function to modify those values.
//
// See https://docs.aws.amazon.com/secretsmanager/latest/userguide/retrieving-secrets_lambda.html.
func Secrets(secretId string, opts ...func(*SecretsOpts)) Variable {
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

	params := SecretsOpts{SecretId: secretId}
	for _, opt := range opts {
		opt(&params)
	}

	client := http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:"+port+"/secretsmanager/get", nil)
	if err != nil {
		return errVar{err: fmt.Errorf("create HTTP request error: %w", err)}
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

	return getter(func() (string, error) {
		res, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("GET secrets manager error: %w", err)
		}

		output := secretsmanager.GetSecretValueOutput{}
		err = json.NewDecoder(res.Body).Decode(&output)
		_ = res.Body.Close()
		if err != nil {
			return "", fmt.Errorf("decode GET secrets manager error: %w", err)
		}

		return aws.ToString(output.SecretString), nil
	})
}
