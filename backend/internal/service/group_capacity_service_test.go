//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type groupCapacityAccountRepoStub struct {
	accounts []Account
	onList   func()
}

func (s *groupCapacityAccountRepoStub) Create(context.Context, *Account) error { return nil }
func (s *groupCapacityAccountRepoStub) GetByID(context.Context, int64) (*Account, error) {
	return nil, ErrAccountNotFound
}
func (s *groupCapacityAccountRepoStub) GetByIDs(context.Context, []int64) ([]*Account, error) {
	return nil, nil
}
func (s *groupCapacityAccountRepoStub) ExistsByID(context.Context, int64) (bool, error) {
	return false, nil
}
func (s *groupCapacityAccountRepoStub) GetByCRSAccountID(context.Context, string) (*Account, error) {
	return nil, nil
}
func (s *groupCapacityAccountRepoStub) FindByExtraField(context.Context, string, any) ([]Account, error) {
	return nil, nil
}
func (s *groupCapacityAccountRepoStub) ListCRSAccountIDs(context.Context) (map[string]int64, error) {
	return nil, nil
}
func (s *groupCapacityAccountRepoStub) Update(context.Context, *Account) error { return nil }
func (s *groupCapacityAccountRepoStub) Delete(context.Context, int64) error    { return nil }
func (s *groupCapacityAccountRepoStub) List(context.Context, pagination.PaginationParams) ([]Account, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (s *groupCapacityAccountRepoStub) ListWithFilters(context.Context, pagination.PaginationParams, string, string, string, string, int64, string) ([]Account, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (s *groupCapacityAccountRepoStub) ListByGroup(context.Context, int64) ([]Account, error) {
	return nil, nil
}
func (s *groupCapacityAccountRepoStub) ListActive(context.Context) ([]Account, error) {
	return nil, nil
}
func (s *groupCapacityAccountRepoStub) ListByPlatform(context.Context, string) ([]Account, error) {
	return nil, nil
}
func (s *groupCapacityAccountRepoStub) UpdateLastUsed(context.Context, int64) error { return nil }
func (s *groupCapacityAccountRepoStub) BatchUpdateLastUsed(context.Context, map[int64]time.Time) error {
	return nil
}
func (s *groupCapacityAccountRepoStub) SetError(context.Context, int64, string) error {
	return nil
}
func (s *groupCapacityAccountRepoStub) ClearError(context.Context, int64) error { return nil }
func (s *groupCapacityAccountRepoStub) SetSchedulable(context.Context, int64, bool) error {
	return nil
}
func (s *groupCapacityAccountRepoStub) AutoPauseExpiredAccounts(context.Context, time.Time) (int64, error) {
	return 0, nil
}
func (s *groupCapacityAccountRepoStub) BindGroups(context.Context, int64, []int64) error {
	return nil
}
func (s *groupCapacityAccountRepoStub) ListSchedulable(context.Context) ([]Account, error) {
	return nil, nil
}
func (s *groupCapacityAccountRepoStub) ListSchedulableByGroupID(ctx context.Context, groupID int64) ([]Account, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s.onList != nil {
		s.onList()
	}
	return s.accounts, nil
}
func (s *groupCapacityAccountRepoStub) ListSchedulableByPlatform(context.Context, string) ([]Account, error) {
	return nil, nil
}
func (s *groupCapacityAccountRepoStub) ListSchedulableByGroupIDAndPlatform(context.Context, int64, string) ([]Account, error) {
	return nil, nil
}
func (s *groupCapacityAccountRepoStub) ListSchedulableByPlatforms(context.Context, []string) ([]Account, error) {
	return nil, nil
}
func (s *groupCapacityAccountRepoStub) ListSchedulableByGroupIDAndPlatforms(context.Context, int64, []string) ([]Account, error) {
	return nil, nil
}
func (s *groupCapacityAccountRepoStub) ListSchedulableUngroupedByPlatform(context.Context, string) ([]Account, error) {
	return nil, nil
}
func (s *groupCapacityAccountRepoStub) ListSchedulableUngroupedByPlatforms(context.Context, []string) ([]Account, error) {
	return nil, nil
}
func (s *groupCapacityAccountRepoStub) SetRateLimited(context.Context, int64, time.Time) error {
	return nil
}
func (s *groupCapacityAccountRepoStub) SetModelRateLimit(context.Context, int64, string, time.Time) error {
	return nil
}
func (s *groupCapacityAccountRepoStub) SetOverloaded(context.Context, int64, time.Time) error {
	return nil
}
func (s *groupCapacityAccountRepoStub) SetTempUnschedulable(context.Context, int64, time.Time, string) error {
	return nil
}
func (s *groupCapacityAccountRepoStub) ClearTempUnschedulable(context.Context, int64) error {
	return nil
}
func (s *groupCapacityAccountRepoStub) ClearRateLimit(context.Context, int64) error { return nil }
func (s *groupCapacityAccountRepoStub) ClearAntigravityQuotaScopes(context.Context, int64) error {
	return nil
}
func (s *groupCapacityAccountRepoStub) ClearModelRateLimits(context.Context, int64) error {
	return nil
}
func (s *groupCapacityAccountRepoStub) UpdateSessionWindow(context.Context, int64, *time.Time, *time.Time, string) error {
	return nil
}
func (s *groupCapacityAccountRepoStub) UpdateExtra(context.Context, int64, map[string]any) error {
	return nil
}
func (s *groupCapacityAccountRepoStub) BulkUpdate(context.Context, []int64, AccountBulkUpdate) (int64, error) {
	return 0, nil
}
func (s *groupCapacityAccountRepoStub) IncrementQuotaUsed(context.Context, int64, float64) error {
	return nil
}
func (s *groupCapacityAccountRepoStub) ResetQuotaUsed(context.Context, int64) error {
	return nil
}

type groupCapacityConcurrencyCacheStub struct {
	counts map[int64]int
}

func (s *groupCapacityConcurrencyCacheStub) AcquireAccountSlot(context.Context, int64, int, string) (bool, error) {
	return false, nil
}
func (s *groupCapacityConcurrencyCacheStub) ReleaseAccountSlot(context.Context, int64, string) error {
	return nil
}
func (s *groupCapacityConcurrencyCacheStub) GetAccountConcurrency(context.Context, int64) (int, error) {
	return 0, nil
}
func (s *groupCapacityConcurrencyCacheStub) GetAccountConcurrencyBatch(ctx context.Context, accountIDs []int64) (map[int64]int, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return s.counts, nil
}
func (s *groupCapacityConcurrencyCacheStub) IncrementAccountWaitCount(context.Context, int64, int) (bool, error) {
	return false, nil
}
func (s *groupCapacityConcurrencyCacheStub) DecrementAccountWaitCount(context.Context, int64) error {
	return nil
}
func (s *groupCapacityConcurrencyCacheStub) GetAccountWaitingCount(context.Context, int64) (int, error) {
	return 0, nil
}
func (s *groupCapacityConcurrencyCacheStub) AcquireUserSlot(context.Context, int64, int, string) (bool, error) {
	return false, nil
}
func (s *groupCapacityConcurrencyCacheStub) ReleaseUserSlot(context.Context, int64, string) error {
	return nil
}
func (s *groupCapacityConcurrencyCacheStub) GetUserConcurrency(context.Context, int64) (int, error) {
	return 0, nil
}
func (s *groupCapacityConcurrencyCacheStub) IncrementWaitCount(context.Context, int64, int) (bool, error) {
	return false, nil
}
func (s *groupCapacityConcurrencyCacheStub) DecrementWaitCount(context.Context, int64) error {
	return nil
}
func (s *groupCapacityConcurrencyCacheStub) GetAccountsLoadBatch(context.Context, []AccountWithConcurrency) (map[int64]*AccountLoadInfo, error) {
	return nil, nil
}
func (s *groupCapacityConcurrencyCacheStub) GetUsersLoadBatch(context.Context, []UserWithConcurrency) (map[int64]*UserLoadInfo, error) {
	return nil, nil
}
func (s *groupCapacityConcurrencyCacheStub) CleanupExpiredAccountSlots(context.Context, int64) error {
	return nil
}
func (s *groupCapacityConcurrencyCacheStub) CleanupStaleProcessSlots(context.Context, string) error {
	return nil
}

func TestGroupCapacityService_RuntimeMetricsIgnoreCanceledRequestContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	repo := &groupCapacityAccountRepoStub{
		accounts: []Account{{ID: 11, Concurrency: 5}, {ID: 12, Concurrency: 7}},
		onList:   cancel,
	}
	cache := &groupCapacityConcurrencyCacheStub{counts: map[int64]int{11: 2, 12: 3}}
	svc := &GroupCapacityService{
		accountRepo:        repo,
		concurrencyService: NewConcurrencyService(cache),
	}

	cap, err := svc.getGroupCapacity(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, 5, cap.ConcurrencyUsed)
	require.Equal(t, 12, cap.ConcurrencyMax)
}
