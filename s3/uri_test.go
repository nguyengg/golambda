package s3

import (
	"reflect"
	"testing"
)

func TestParseS3URIWithOwner(t *testing.T) {
	type args struct {
		rawURL string
	}
	tests := []struct {
		name      string
		args      args
		wantValue URIWithOwner
		wantErr   bool
	}{
		// TODO: Add test cases.
		{
			name: "S3 URI to a file",
			args: args{rawURL: "s3://my-bucket[1234]/path/to/file.json"},
			wantValue: URIWithOwner{
				Bucket:              "my-bucket",
				Key:                 "path/to/file.json",
				ExpectedBucketOwner: "1234",
			},
		},
		{
			name: "S3 prefix path",
			args: args{rawURL: "s3://my-bucket[1234]/path/to/"},
			wantValue: URIWithOwner{
				Bucket:              "my-bucket",
				Key:                 "path/to/",
				ExpectedBucketOwner: "1234",
			},
		},
		{
			name: "optional S3 key",
			args: args{rawURL: "s3://my-bucket[1234]/"},
			wantValue: URIWithOwner{
				Bucket:              "my-bucket",
				Key:                 "",
				ExpectedBucketOwner: "1234",
			},
		},
		{
			name: "optional S3 key without trailing slash",
			args: args{rawURL: "s3://my-bucket[1234]"},
			wantValue: URIWithOwner{
				Bucket:              "my-bucket",
				Key:                 "",
				ExpectedBucketOwner: "1234",
			},
		},
		{
			name:    "not an S3 URI",
			args:    args{rawURL: "https://smile.amazon.com"},
			wantErr: true,
		},
		{
			name:    "missing bucket owner",
			args:    args{rawURL: "s3://my-bucket/path/to/file.json"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, err := ParseURIWithOwner(tt.args.rawURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseURIWithOwner() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotValue, tt.wantValue) {
				t.Errorf("ParseURIWithOwner() gotValue = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}
