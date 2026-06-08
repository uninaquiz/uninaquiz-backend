package mappers_test

import (
	"testing"
	"time"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/commands"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/mappers"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/entities"
)

func TestToUserResponse(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name  string
		input entities.User
	}{
		{
			name: "maps all fields correctly",
			input: entities.User{
				ID:        "uuid-1",
				Username:  "alice",
				Password:  "should-not-appear",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		{
			name:  "zero-value user",
			input: entities.User{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := mappers.ToUserResponse(tt.input)
			if resp == nil {
				t.Fatal("ToUserResponse() returned nil")
			}
			if resp.ID != tt.input.ID {
				t.Errorf("ID = %v, want %v", resp.ID, tt.input.ID)
			}
			if resp.Username != tt.input.Username {
				t.Errorf("Username = %v, want %v", resp.Username, tt.input.Username)
			}
			if !resp.CreatedAt.Equal(tt.input.CreatedAt) {
				t.Errorf("CreatedAt = %v, want %v", resp.CreatedAt, tt.input.CreatedAt)
			}
			if !resp.UpdatedAt.Equal(tt.input.UpdatedAt) {
				t.Errorf("UpdatedAt = %v, want %v", resp.UpdatedAt, tt.input.UpdatedAt)
			}
		})
	}
}

func TestToUserEntity(t *testing.T) {
	tests := []struct {
		name  string
		input commands.CreateUserCommand
	}{
		{
			name:  "maps username and password, generates UUID",
			input: commands.CreateUserCommand{Username: "bob", Password: "pass1234"},
		},
		{
			name:  "different credentials produce unique IDs",
			input: commands.CreateUserCommand{Username: "carol", Password: "password99"},
		},
	}

	seenIDs := map[string]bool{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := mappers.ToUserEntity(tt.input)
			if entity == nil {
				t.Fatal("ToUserEntity() returned nil")
			}
			if entity.Username != tt.input.Username {
				t.Errorf("Username = %v, want %v", entity.Username, tt.input.Username)
			}
			if entity.Password != tt.input.Password {
				t.Errorf("Password = %v, want %v", entity.Password, tt.input.Password)
			}
			if entity.ID == "" {
				t.Error("ID should be a non-empty UUID")
			}
			if seenIDs[entity.ID] {
				t.Errorf("duplicate UUID generated: %v", entity.ID)
			}
			seenIDs[entity.ID] = true
		})
	}
}
