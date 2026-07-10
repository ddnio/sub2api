package service

import (
	"context"
	"strconv"
)

func (s *SettingService) IsReferralEnabled(ctx context.Context) bool {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyReferralEnabled)
	if err != nil {
		return false
	}
	return value == "true"
}

func (s *SettingService) GetReferralInviterAmount(ctx context.Context) float64 {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyReferralInviterAmount)
	if err != nil {
		return 0
	}
	if v, err := strconv.ParseFloat(value, 64); err == nil && v >= 0 {
		return v
	}
	return 0
}

func (s *SettingService) GetReferralInviteeAmount(ctx context.Context) float64 {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyReferralInviteeAmount)
	if err != nil {
		return 0
	}
	if v, err := strconv.ParseFloat(value, 64); err == nil && v >= 0 {
		return v
	}
	return 0
}
