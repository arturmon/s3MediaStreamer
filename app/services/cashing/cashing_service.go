package cashing

import (
	"context"
)

type CachingRepository interface {
	CachePasswordVerificationInRedis(ctx context.Context, userID string, verificationSuccess bool, expiration int) error
	CheckPasswordVerificationInRedis(ctx context.Context, userID string) (bool, bool, error)
}

type CachingService struct {
	redisRepository CachingRepository
}

func NewCachingService(redisRepository CachingRepository) *CachingService {
	return &CachingService{redisRepository: redisRepository}
}

func (s *CachingService) CachePasswordVerificationInRedis(ctx context.Context, userID string, verificationSuccess bool, expiration int) error {
	return s.redisRepository.CachePasswordVerificationInRedis(ctx, userID, verificationSuccess, expiration)
}

func (s *CachingService) CheckPasswordVerificationInRedis(ctx context.Context, userID string) (bool, bool, error) {
	return s.redisRepository.CheckPasswordVerificationInRedis(ctx, userID)
}
