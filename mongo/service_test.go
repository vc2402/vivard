package mongo

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		params []any
	}
	tests := []struct {
		name    string
		args    args
		want    *Service
		wantErr bool
	}{
		{
			name: "dbName only",
			args: args{[]any{"dbName"}},
			want: &Service{
				config: map[string]ConnectionConfig{"default": {Alias: "default", DBName: "dbName"}},
			},
		},
		{
			name: "connect string with dbName",
			args: args{[]any{"connectString", "dbName"}},
			want: &Service{
				config: map[string]ConnectionConfig{"default": {Alias: "default", ConnectString: "connectString", DBName: "dbName"}},
			},
		},
		{
			name: "connect string with dbName and ConnectionConfig",
			args: args{[]any{
				"connectString", "db name",
				ConnectionConfig{Alias: "another", ConnectString: "someString", DBName: "someName"},
			}},
			want: &Service{
				config: map[string]ConnectionConfig{
					"default": {Alias: "default", ConnectString: "connectString", DBName: "db name"},
					"another": {Alias: "another", ConnectString: "someString", DBName: "someName"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.params...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() got = %v, want %v", got, tt.want)
			}
		})
	}
}
