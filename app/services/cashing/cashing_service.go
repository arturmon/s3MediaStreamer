package cashing

import (
	"context"
	"time"
)

type CachingRepository interface {
	CachePasswordVerificationInRedis(ctx context.Context, userID string, verificationSuccess bool, expiration int) error
	CheckPasswordVerificationInRedis(ctx context.Context, userID string) (bool, bool, error)
	GetTrackInCache(ctx context.Context, key string, result interface{}) error
	SetTrackInCache(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	UpdateTrackInCache(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	DeleteTrackInCache(ctx context.Context, key string) error
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

func (s *CachingService) GetTrackInCache(ctx context.Context, key string, result interface{}) error {
	return s.redisRepository.GetTrackInCache(ctx, key, result)
}

func (s *CachingService) SetTrackInCache(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return s.redisRepository.SetTrackInCache(ctx, key, value, ttl)
}

func (s *CachingService) UpdateTrackInCache(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return s.redisRepository.UpdateTrackInCache(ctx, key, value, ttl)
}

func (s *CachingService) DeleteTrackInCache(ctx context.Context, key string) error {
	return s.redisRepository.DeleteTrackInCache(ctx, key)
}
