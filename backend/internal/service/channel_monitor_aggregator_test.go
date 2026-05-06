package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilterUserVisibleMonitorsUsesAllowedGroupNames(t *testing.T) {
	allowed := map[string]struct{}{
		"public": {},
	}
	monitors := []*ChannelMonitor{
		{ID: 1, GroupName: "public", Enabled: true},
		{ID: 2, GroupName: "private", Enabled: true},
		{ID: 3, GroupName: " public ", Enabled: true},
		{ID: 4, GroupName: "", Enabled: true},
		{ID: 5, GroupName: "public", Enabled: false},
		nil,
	}

	visible := filterUserVisibleMonitors(monitors, allowed)

	require.Len(t, visible, 2)
	require.Equal(t, int64(1), visible[0].ID)
	require.Equal(t, int64(3), visible[1].ID)
}

func TestMonitorVisibleToGroupNamesRequiresExplicitAllowedGroup(t *testing.T) {
	require.False(t, monitorVisibleToGroupNames(&ChannelMonitor{GroupName: "public", Enabled: true}, nil))
	require.False(t, monitorVisibleToGroupNames(&ChannelMonitor{GroupName: "public", Enabled: true}, map[string]struct{}{}))
	require.False(t, monitorVisibleToGroupNames(&ChannelMonitor{GroupName: "", Enabled: true}, map[string]struct{}{"": {}}))
	require.False(t, monitorVisibleToGroupNames(&ChannelMonitor{GroupName: "Public", Enabled: true}, map[string]struct{}{"public": {}}))
	require.True(t, monitorVisibleToGroupNames(&ChannelMonitor{GroupName: "public", Enabled: true}, map[string]struct{}{"public": {}}))
}
