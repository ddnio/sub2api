package repository

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEscapeImageCreationLike(t *testing.T) {
	require.Equal(t, `50\%\_off\\today`, escapeImageCreationLike(`50%_off\today`))
}
