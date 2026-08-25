package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIsImageCreationAPIKeyEligible(t *testing.T) {
	groupID := int64(7)
	expiredAt := time.Now().Add(-time.Minute)
	activeGroup := &Group{ID: groupID, Status: StatusActive, AllowImageGeneration: true}
	allowedUser := &User{AllowedGroups: []int64{groupID}}

	tests := []struct {
		name string
		key  *APIKey
		want bool
	}{
		{name: "nil", want: false},
		{name: "ungrouped active", key: &APIKey{Status: StatusActive}, want: true},
		{name: "disabled key", key: &APIKey{Status: StatusAPIKeyDisabled}, want: false},
		{name: "expired key", key: &APIKey{Status: StatusActive, ExpiresAt: &expiredAt}, want: false},
		{name: "quota exhausted", key: &APIKey{Status: StatusActive, Quota: 1, QuotaUsed: 1}, want: false},
		{name: "missing group", key: &APIKey{Status: StatusActive, GroupID: &groupID}, want: false},
		{name: "disabled group", key: &APIKey{Status: StatusActive, GroupID: &groupID, Group: &Group{ID: groupID, Status: StatusDisabled, AllowImageGeneration: true}}, want: false},
		{name: "deleted group", key: &APIKey{Status: StatusActive, GroupID: &groupID, Group: &Group{ID: groupID, Status: "deleted", AllowImageGeneration: true}}, want: false},
		{name: "image generation disabled", key: &APIKey{Status: StatusActive, GroupID: &groupID, Group: &Group{ID: groupID, Status: StatusActive}}, want: false},
		{name: "public group", key: &APIKey{Status: StatusActive, GroupID: &groupID, Group: activeGroup, User: &User{}}, want: true},
		{name: "exclusive group allowed", key: &APIKey{Status: StatusActive, GroupID: &groupID, Group: &Group{ID: groupID, Status: StatusActive, IsExclusive: true, AllowImageGeneration: true}, User: allowedUser}, want: true},
		{name: "exclusive group revoked", key: &APIKey{Status: StatusActive, GroupID: &groupID, Group: &Group{ID: groupID, Status: StatusActive, IsExclusive: true, AllowImageGeneration: true}, User: &User{}}, want: false},
		{name: "subscription group", key: &APIKey{Status: StatusActive, GroupID: &groupID, Group: &Group{ID: groupID, Status: StatusActive, SubscriptionType: SubscriptionTypeSubscription, AllowImageGeneration: true}, User: &User{}}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, IsImageCreationAPIKeyEligible(tt.key))
		})
	}
}
