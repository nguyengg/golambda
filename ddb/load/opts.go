package load

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nguyengg/golambda/ddb/expr"
	"github.com/nguyengg/golambda/ddb/model"
)

// Opts provides customisation to the dynamodb.GetItemInput made with [github.com/nguyengg/golambda/ddb.Wrapper.Load].
//
// Opts.Input is guaranteed to exist when passed into the first modifier.
// Opts.Item is the original reference item. Changes to Opts.Item don't automatically update Opts.Input.
// Changes to Opts.Projection will affect the final Opts.Input.
type Opts struct {
	Item       model.Item
	Input      *dynamodb.GetItemInput
	Projection *expression.ProjectionBuilder
}

// WithProjection adds a projection expression.
func WithProjection(name string, names ...string) func(*Opts) {
	return func(opts *Opts) {
		switch len(names) {
		case 0:
			opts.Projection = expr.AddNames(opts.Projection, expression.Name(name))
		case 1:
			opts.Projection = expr.AddNames(opts.Projection, expression.Name(name), expression.Name(names[0]))
		case 2:
			opts.Projection = expr.AddNames(opts.Projection, expression.Name(name), expression.Name(names[0]), expression.Name(names[1]))
		default:
			nameBuilders := make([]expression.NameBuilder, len(names))
			for i, n := range names {
				nameBuilders[i] = expression.Name(n)
			}
			opts.Projection = expr.AddNames(opts.Projection, expression.Name(name), nameBuilders...)
		}
	}
}
