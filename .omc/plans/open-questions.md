# Open Questions

## referral-feature - 2026-04-08

- [ ] **Existing user backfill**: Should existing users get `referral_code` generated retroactively (via a one-time migration/script), or only on next login/profile visit? -- Affects whether existing users can immediately share referral links.
- [ ] **RegisterWithVerification signature change**: Adding `referralCode` as a 7th string parameter is getting unwieldy. Consider refactoring to an options struct (e.g., `RegisterOptions{PromoCode, InvitationCode, ReferralCode}`) -- but that's a larger refactor. Executor should decide: add parameter vs. introduce struct.
- [ ] **Frontend field reuse vs. separate field**: The `invitation_code` JSON field in RegisterRequest can be reused (backend interprets based on mode), or a new `referral_code` field can be added. Reuse is simpler but muddier semantically. Executor should pick the cleaner approach.
- [ ] **OAuth registration path**: `LoginOrRegisterOAuthWithTokenPair` also handles registration (LinuxDo OAuth). Should it also support referral codes? Currently the OAuth callback flow doesn't pass referral codes. This could be a follow-up feature.
- [ ] **Admin: referral code regeneration**: Should admins be able to regenerate a user's referral code? Not in v1 scope, but worth noting for future.
- [ ] **Rate limiting on referral rewards**: Should there be a cap on how many referrals one user can make (e.g., max 100 invitees)? Not in v1 scope, but could be exploited if reward amounts are high.
