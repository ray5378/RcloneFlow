package service

import "context"

// JobSyncService 已退役：运行态主链已切到 CLI runner 收尾，不再通过 RC job/status 轮询同步。
type JobSyncService struct{}

func NewJobSyncService(_ any, _ any, _ int) *JobSyncService { return &JobSyncService{} }
func (s *JobSyncService) Start(_ context.Context)           {}
func (s *JobSyncService) Stop()                            {}
