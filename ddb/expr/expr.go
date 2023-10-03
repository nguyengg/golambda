// Package expr provides utility methods to manipulate expression builders from [github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression].

package expr

import "github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"

// AddNames adds names to the nillable "left" expression.ProjectionBuilder.
//
// Return non-nil.
func AddNames(left *expression.ProjectionBuilder, nameBuilder expression.NameBuilder, nameBuilders ...expression.NameBuilder) *expression.ProjectionBuilder {
	if left == nil {
		p := expression.NamesList(nameBuilder, nameBuilders...)
		return &p
	} else {
		nameBuilders = append([]expression.NameBuilder{nameBuilder}, nameBuilders...)
		p := expression.AddNames(*left, nameBuilders...)
		return &p
	}
}

// And adds an AND condition expression to the nillable "left" expression.ConditionBuilder.
//
// Return non-nil.
func And(left *expression.ConditionBuilder, right expression.ConditionBuilder, other ...expression.ConditionBuilder) *expression.ConditionBuilder {
	if left == nil {
		if len(other) == 0 {
			return &right
		} else {
			cond := right.And(other[0], other[1:]...)
			return &cond
		}
	} else {
		cond := left.And(right, other...)
		return &cond
	}
}

// Or adds an OR condition expression to the nillable "left" expression.ConditionBuilder.
//
// Return non-nil.
func Or(left *expression.ConditionBuilder, right expression.ConditionBuilder, other ...expression.ConditionBuilder) *expression.ConditionBuilder {
	if left == nil {
		if len(other) == 0 {
			return &right
		} else {
			cond := right.Or(other[0], other[1:]...)
			return &cond
		}
	} else {
		cond := left.Or(right, other...)
		return &cond
	}
}

// Add adds an ADD update expression to the nillable "left" expression.UpdateBuilder.
//
// Return non-nil.
func Add(left *expression.UpdateBuilder, name expression.NameBuilder, value expression.ValueBuilder) *expression.UpdateBuilder {
	if left == nil {
		expr := expression.Add(name, value)
		return &expr
	} else {
		expr := left.Add(name, value)
		return &expr
	}
}

// Set adds a SET update expression to the nillable "left" expression.UpdateBuilder.
//
// Return non-nil.
func Set(left *expression.UpdateBuilder, name expression.NameBuilder, value expression.ValueBuilder) *expression.UpdateBuilder {
	if left == nil {
		expr := expression.Set(name, value)
		return &expr
	} else {
		expr := left.Set(name, value)
		return &expr
	}
}

// Delete adds a DELETE update expression to the nillable "left" expression.UpdateBuilder.
//
// Return non-nil.
func Delete(left *expression.UpdateBuilder, name expression.NameBuilder, value expression.ValueBuilder) *expression.UpdateBuilder {
	if left == nil {
		expr := expression.Delete(name, value)
		return &expr
	} else {
		expr := left.Delete(name, value)
		return &expr
	}
}

// Remove adds a Remove update expression to the nillable "left" expression.UpdateBuilder.
//
// Return non-nil.
func Remove(left *expression.UpdateBuilder, name expression.NameBuilder) *expression.UpdateBuilder {
	if left == nil {
		expr := expression.Remove(name)
		return &expr
	} else {
		expr := left.Remove(name)
		return &expr
	}
}
