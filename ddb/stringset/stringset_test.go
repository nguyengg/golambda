package stringset

import "testing"

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
