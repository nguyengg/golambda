package etag

import (
	"reflect"
	"testing"
)

func TestParseDirectives(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name    string
		args    args
		want    *Directives
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "empty value",
			args: args{value: ""},
			want: nil,
		},
		{
			name: "any",
			args: args{value: "*"},
			want: &Directives{Any: true},
		},
		{
			name: "single strong value",
			args: args{value: "\"bfc13a64729c4290ef5b2c2730249c88ca92d82d\""},
			want: &Directives{ETags: []ETag{NewStrongETag("bfc13a64729c4290ef5b2c2730249c88ca92d82d")}},
		},
		{
			name: "mix and match",
			args: args{value: "W/\"67ab43\", \"54ed21\", W/\"7892dd\""},
			want: &Directives{ETags: []ETag{
				NewWeakETag("67ab43"),
				NewStrongETag("54ed21"),
				NewWeakETag("7892dd"),
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDirectives(tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDirectives() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseDirectives() got = %v, want %v", got, tt.want)
			}
		})
	}
}
