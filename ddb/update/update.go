package update

import "github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"

// Add adds an ADD update expression to the nil-safe update expression.UpdateBuilder.
// Returns non-nil expression.ConditionBuilder.
func Add(update *expression.UpdateBuilder, name expression.NameBuilder, value expression.ValueBuilder) *expression.UpdateBuilder {
	if update == nil {
		expr := expression.Add(name, value)
		return &expr
	} else {
		expr := update.Add(name, value)
		return &expr
	}
}

// Set adds a SET update expression to the nil-safe update expression.UpdateBuilder.
// Returns non-nil expression.ConditionBuilder.
func Set(update *expression.UpdateBuilder, name expression.NameBuilder, value expression.ValueBuilder) *expression.UpdateBuilder {
	if update == nil {
		expr := expression.Set(name, value)
		return &expr
	} else {
		expr := update.Set(name, value)
		return &expr
	}
}

// Delete adds a DELETE update expression to the nil-safe update expression.UpdateBuilder.
// Returns non-nil expression.ConditionBuilder.
func Delete(update *expression.UpdateBuilder, name expression.NameBuilder, value expression.ValueBuilder) *expression.UpdateBuilder {
	if update == nil {
		expr := expression.Delete(name, value)
		return &expr
	} else {
		expr := update.Delete(name, value)
		return &expr
	}
}

// Remove adds an Remove update expression to the nil-safe update expression.UpdateBuilder.
// Returns non-nil expression.ConditionBuilder.
func Remove(update *expression.UpdateBuilder, name expression.NameBuilder) *expression.UpdateBuilder {
	if update == nil {
		expr := expression.Remove(name)
		return &expr
	} else {
		expr := update.Remove(name)
		return &expr
	}
}
