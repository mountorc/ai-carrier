package sql

import (
	"testing"
)

func TestGetSQL(t *testing.T) {
	tests := []struct {
		name    string
		uuid    string
		where   map[string]interface{}
		wantErr bool
	}{
		{
			name: "listSkills without where",
			uuid: "330e8400-e22b-41d4-a716-446655440001",
			where: nil,
			wantErr: false,
		},
		{
			name: "listSkills with vectorText",
			uuid: "330e8400-e22b-41d4-a716-446655440001",
			where: map[string]interface{}{
				"vectorText": "test query",
			},
			wantErr: false,
		},
		{
			name: "getSkill",
			uuid: "330e8400-e29b-41d4-a716-446655440002",
			where: nil,
			wantErr: false,
		},
		{
			name:    "invalid uuid",
			uuid:    "invalid-uuid",
			where:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetSQL(tt.uuid, tt.where)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSQL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				t.Logf("Generated SQL: %s", result)
			}
		})
	}
}
