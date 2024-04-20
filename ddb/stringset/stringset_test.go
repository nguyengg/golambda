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

func TestStringSet_Delete(t *testing.T) {
	type args struct {
		value     string
		lenBefore int
		lenAfter  int
	}
	tests := []struct {
		name   string
		m      StringSet
		args   args
		wantOk bool
	}{
		// TODO: Add test cases.
		{
			name: "delete OK",
			m:    []string{"a", "b", "c"},
			args: args{
				value:     "a",
				lenBefore: 3,
				lenAfter:  2,
			},
			wantOk: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if lenBefore := len(tt.m); lenBefore != tt.args.lenBefore {
				t.Errorf("lenBefore = %v, want %v", lenBefore, tt.args.lenBefore)
			}

			if gotOk := tt.m.Delete(tt.args.value); gotOk != tt.wantOk {
				t.Errorf("Delete() = %v, want %v", gotOk, tt.wantOk)
			}

			if tt.wantOk && tt.m.Has(tt.args.value) {
				t.Errorf("Has() = %t, want %t", true, false)
			}

			if lenAfter := len(tt.m); lenAfter != tt.args.lenAfter {
				t.Errorf("lenAfter = %v, want %v", lenAfter, tt.args.lenAfter)
			}
		})
	}
}
