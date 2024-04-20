package stringset

import (
	"reflect"
	"testing"
)

func TestStringSet_IsSubset(t *testing.T) {
	type args struct {
		other StringSet
	}
	tests := []struct {
		name string
		m    StringSet
		args args
		want bool
	}{
		{
			name: "is subset",
			m:    []string{"a", "b"},
			args: args{other: []string{"a", "b", "c"}},
			want: true,
		},
		{
			name: "is not subset",
			m:    []string{"a", "b", "c"},
			args: args{other: []string{"a", "b", "d"}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.IsSubset(tt.args.other); got != tt.want {
				t.Errorf("IsSubset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringSet_IsSuperset(t *testing.T) {
	type args struct {
		other StringSet
	}
	tests := []struct {
		name string
		m    StringSet
		args args
		want bool
	}{
		{
			name: "is superset",
			m:    []string{"a", "b", "c"},
			args: args{other: []string{"a", "b"}},
			want: true,
		},
		{
			name: "is not superset",
			m:    []string{"a", "b", "c"},
			args: args{other: []string{"a", "b", "d"}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.IsSuperset(tt.args.other); got != tt.want {
				t.Errorf("IsSuperset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringSet_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		m       StringSet
		want    []byte
		wantErr bool
	}{
		{
			name: "marshal",
			m:    []string{"a", "b", "c"},
			want: []byte(`["a","b","c"]`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalJSON() got = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestStringSet_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		m       StringSet
		args    args
		wantErr bool
	}{
		{
			name: "unmarshal",
			m:    []string{"a", "b", "c"},
			args: args{data: []byte(`["a","b","c"]`)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.m.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
