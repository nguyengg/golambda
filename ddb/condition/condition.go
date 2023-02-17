package condition

import "github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"

// Adds an AND condition expression to the possibly nil condition expression.ConditionBuilder.
// Return non-nll expression.ConditionBuilder.
func And(condition *expression.ConditionBuilder, right expression.ConditionBuilder, other ...expression.ConditionBuilder) *expression.ConditionBuilder {
	if condition == nil {
		if len(other) == 0 {
			return &right
		} else {
			cond := right.And(other[0], other[1:]...)
			return &cond
		}
	} else {
		cond := condition.And(right, other...)
		return &cond
	}
}

// Adds an OR condition expression to the possibly nil condition expression.ConditionBuilder.
// Return non-nil expression.ConditionBuilder.
func Or(condition *expression.ConditionBuilder, right expression.ConditionBuilder, other ...expression.ConditionBuilder) *expression.ConditionBuilder {
	if condition == nil {
		if len(other) == 0 {
			return &right
		} else {
			cond := right.Or(other[0], other[1:]...)
			return &cond
		}
	} else {
		cond := condition.Or(right, other...)
		return &cond
	}
}
