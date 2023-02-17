package auth

import "github.com/aws/aws-lambda-go/events"

// APIGatewayHTTPLambdaAuthorizerV2Request is the request payload version 2.0 from
// https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-lambda-authorizer.html.
type APIGatewayHTTPLambdaAuthorizerV2Request struct {
	Version               string                                         `json:"version"`
	Type                  string                                         `json:"type"`
	RouteArn              string                                         `json:"routeArn"`
	IdentitySource        []string                                       `json:"identitySource,omitempty"`
	RouteKey              string                                         `json:"routeKey"`
	RawPath               string                                         `json:"rawPath"`
	RawQueryString        string                                         `json:"rawQueryString,omitempty"`
	Cookies               []string                                       `json:"cookies,omitempty"`
	Headers               map[string]string                              `json:"headers"`
	QueryStringParameters map[string]string                              `json:"queryStringParameters,omitempty"`
	RequestContext        APIGatewayHTTPLambdaAuthorizerV2RequestContext `json:"requestContext"`
	PathParameters        map[string]string                              `json:"pathParameters,omitempty"`
	StageVariables        map[string]string                              `json:"stageVariables,omitempty"`
}

type APIGatewayHTTPLambdaAuthorizerV2RequestContext struct {
	AccountID      string                                                        `json:"accountId"`
	AppID          string                                                        `json:"appId"`
	Authentication *APIGatewayHTTPLambdaAuthorizerV2RequestContextAuthentication `json:"authentication,omitempty"`
	DomainName     string                                                        `json:"domainName"`
	DomainPrefix   string                                                        `json:"domainPrefix"`
	HTTP           events.APIGatewayV2HTTPRequestContextHTTPDescription          `json:"http"`
	RequestID      string                                                        `json:"requestId"`
	RouteKey       string                                                        `json:"routeKey"`
	Stage          string                                                        `json:"stage"`
	Time           string                                                        `json:"time"`
	TimeEpoch      int64                                                         `json:"timeEpoch"`
}

type APIGatewayHTTPLambdaAuthorizerV2RequestContextAuthentication struct {
	ClientCert APIGatewayHTTPLambdaAuthorizerV2RequestContextAuthenticationClientCert `json:"clientCert"`
}

type APIGatewayHTTPLambdaAuthorizerV2RequestContextAuthenticationClientCert struct {
	ClientCertPerm string                                                                         `json:"clientCertPem"`
	SubjectDN      string                                                                         `json:"subjectDN"`
	IssuerDN       string                                                                         `json:"issuerDN"`
	SerialNumber   string                                                                         `json:"serialNumber"`
	Validity       APIGatewayHTTPLambdaAuthorizerV2RequestContextAuthenticationClientCertValidity `json:"validity"`
}

type APIGatewayHTTPLambdaAuthorizerV2RequestContextAuthenticationClientCertValidity struct {
	NotBefore string `json:"notBefore"`
	NotAfter  string `json:"notAfter"`
}

// APIGatewayHTTPLambdaAuthorizerV2Request is the request payload version 2.0 from
// https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-lambda-authorizer.html.
type APIGatewayHTTPLambdaAuthorizerV2SimpleResponse struct {
	IsAuthorized bool              `json:"isAuthorized"`
	Context      map[string]string `json:"context,omitempty"`
}
