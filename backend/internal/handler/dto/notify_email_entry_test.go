package dto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNotifyEmailEntriesFromService_NilEntriesReturnEmptySlice(t *testing.T) {
	entries := NotifyEmailEntriesFromService(nil)
	require.NotNil(t, entries)

	data, err := json.Marshal(entries)
	require.NoError(t, err)
	require.JSONEq(t, `[]`, string(data))
}
