package app

import "time"

const requestCleanupAfterApproval = 7 * 24 * time.Hour

func (s *Server) startRequestCleanupJob() {
	if s.settings.RequestCleanupInterval <= 0 {
		return
	}
	ticker := time.NewTicker(s.settings.RequestCleanupInterval)
	go func() {
		for range ticker.C {
			s.cleanupOldRequests()
		}
	}()
}

func (s *Server) cleanupOldRequests() {
	cutoff := time.Now().Add(-requestCleanupAfterApproval)
	updated, err := s.store.CleanupRequests(cutoff)
	if err != nil {
		s.logger.Printf("cleanup requests failed: %v", err)
		return
	}
	if updated > 0 {
		s.logger.Printf("cleanup requests updated %d rows", updated)
	}
}
