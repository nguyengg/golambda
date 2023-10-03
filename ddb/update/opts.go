package update

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nguyengg/golambda/ddb/expr"
	"github.com/nguyengg/golambda/ddb/model"
	"github.com/nguyengg/golambda/ddb/timestamp"
)

// Opts provides customisation to the dynamodb.UpdateItemInput made with [github.com/nguyengg/golambda/ddb.Wrapper.Update].
//
// Opts.Input is guaranteed to exist when passed into the first modifier.
// Opts.Item is the original reference item. Changes to Opts.Item don't automatically update Opts.Input.
// Changes to Opts.Condition and Opts.Update will affect the final Opts.Input.
type Opts struct {
	Item                          model.Item
	Input                         *dynamodb.UpdateItemInput
	Condition                     *expression.ConditionBuilder
	Update                        *expression.UpdateBuilder
	DisableOptimisticLocking      bool
	DisableAutoGenerateTimestamps timestamp.AutoGenerateFlag
}

// DisableOptimisticLocking disables logic around [model.Versioned].
func DisableOptimisticLocking() func(*Opts) {
	return func(opts *Opts) {
		opts.DisableOptimisticLocking = true
	}
}

// DisableAutoGenerateTimestamps disables logic around [model.HasCreatedTimestamp] and [model.HasModifiedTimestamp].
func DisableAutoGenerateTimestamps(flag timestamp.AutoGenerateFlag) func(*Opts) {
	return func(opts *Opts) {
		opts.DisableAutoGenerateTimestamps = flag
	}
}

// SetOrRemove adds either Set or Remove action to the update expression.
//
// If set is true, a SET action will be added.
// If set is false, only when remove is true then a REMOVE action will be added.
//
// | set   | remove | action
// | true  | *      | SET
// | false | true   | REMOVE
// | false | false  | no-op
//
// This is useful for distinguishing between PUT/POST (remove=true) that replaces attributes with clobbering behaviour
// vs. PATCH (remove=false) that will only update attributes that are non-nil. An example is given:
//
//	func PUT(body Struct) {
//		// because it's a PUT request, if notes is empty, instead of writing empty string to database, we'll remove it.
//		wrapper.Update(..., SetOrRemove(true, true, "notes", body.Notes))
//	}
//
//	func PATCH(body Struct) {
//		// only when notes is non-empty that we'll update it. an empty notes just means caller didn't try to update it.
//		wrapper.Update(..., opts.SetOrRemove(body.Notes != "", false, "notes", body.Notes))
//	}
//
//	func Update(method string, body Struct) {
//		// an attempt to unify the methods may look like this.
//		wrapper.Update(..., opts.SetOrRemove(body.Notes != "", method != "PATCH", "notes", body.Notes))
//	}
//
// If you have multiple actions to be added to the update expression, use this function. If you only have one or two
// actions, you can use the global SetOrRemove as well.
func (opts *Opts) SetOrRemove(set, remove bool, name string, value interface{}) *Opts {
	if set {
		opts.Update = expr.Set(opts.Update, expression.Name(name), expression.Value(value))
		return opts
	}

	if remove {
		opts.Update = expr.Remove(opts.Update, expression.Name(name))
		return opts
	}

	return opts
}

// SetOrRemove is the global variant of Opts.SetOrRemove.
//
// This is useful if you only have one or two actions in the update expression and don't need chaining.
func SetOrRemove(set, remove bool, name string, value interface{}) func(*Opts) {
	return func(opts *Opts) {
		opts.SetOrRemove(set, remove, name, value)
	}
}

// Add is a helper method to expr.Add to add an ADD action.
func (opts *Opts) Add(name string, value interface{}) *Opts {
	opts.Update = expr.Add(opts.Update, expression.Name(name), expression.Value(value))
	return opts
}

// Add is the global variant of Opts.Add.
//
// This is useful if you only have one or two actions in the update expression and don't need chaining.
func Add(name string, value interface{}) func(*Opts) {
	return func(opts *Opts) {
		opts.Add(name, value)
	}
}

// Set is a helper method to expr.Set to add a SET action.
func (opts *Opts) Set(name string, value interface{}) *Opts {
	opts.Update = expr.Set(opts.Update, expression.Name(name), expression.Value(value))
	return opts
}

// Set is the global variant of Opts.Set.
//
// This is useful if you only have one or two actions in the update expression and don't need chaining.
func Set(name string, value interface{}) func(*Opts) {
	return func(opts *Opts) {
		opts.Set(name, value)
	}
}

// Delete is a helper method to expr.Delete to add a DELETE action.
func (opts *Opts) Delete(name string, value interface{}) *Opts {
	opts.Update = expr.Delete(opts.Update, expression.Name(name), expression.Value(value))
	return opts
}

// Delete is the global variant of Opts.Delete.
//
// This is useful if you only have one or two actions in the update expression and don't need chaining.
func Delete(name string, value interface{}) func(*Opts) {
	return func(opts *Opts) {
		opts.Delete(name, value)
	}
}

// Remove is a helper method to expr.Remove to add a REMOVE action.
func (opts *Opts) Remove(name string, value interface{}) *Opts {
	opts.Update = expr.Remove(opts.Update, expression.Name(name))
	return opts
}

// Remove is the global variant of Opts.Remove.
//
// This is useful if you only have one or two actions in the update expression and don't need chaining.
func Remove(name string, value interface{}) func(*Opts) {
	return func(opts *Opts) {
		opts.Remove(name, value)
	}
}

// ReturnAllOldValues sets the dynamodb.UpdateItemInput's ReturnValues to ALL_OLD.
func ReturnAllOldValues() func(*Opts) {
	return func(opts *Opts) {
		opts.Input.ReturnValues = types.ReturnValueAllOld
	}
}

// ReturnUpdatedOldValues sets the dynamodb.UpdateItemInput's ReturnValues to UPDATED_OLD.
func ReturnUpdatedOldValues() func(*Opts) {
	return func(opts *Opts) {
		opts.Input.ReturnValues = types.ReturnValueUpdatedOld
	}
}

// ReturnAllNewValues sets the dynamodb.UpdateItemInput's ReturnValues to ALL_NEW.
func ReturnAllNewValues() func(*Opts) {
	return func(opts *Opts) {
		opts.Input.ReturnValues = types.ReturnValueAllNew
	}
}

// ReturnUpdatedNewValues sets the dynamodb.UpdateItemInput's ReturnValues to UPDATED_NEW.
func ReturnUpdatedNewValues() func(*Opts) {
	return func(opts *Opts) {
		opts.Input.ReturnValues = types.ReturnValueUpdatedNew
	}
}
