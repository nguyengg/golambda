package projection

import "github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"

// Adds names to the possibly nil projection expression.ProjectionBuilder.
// Return non-nil expression.ProjectionBuilder.
func AddNames(projection *expression.ProjectionBuilder, nameBuilder expression.NameBuilder, nameBuilders ...expression.NameBuilder) *expression.ProjectionBuilder {
	if projection == nil {
		p := expression.NamesList(nameBuilder, nameBuilders...)
		return &p
	} else {
		nameBuilders = append([]expression.NameBuilder{nameBuilder}, nameBuilders...)
		p := expression.AddNames(*projection, nameBuilders...)
		return &p
	}
}
