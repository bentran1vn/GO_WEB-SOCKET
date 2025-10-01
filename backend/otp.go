package main

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type OTP struct {
	Key     string
	Created time.Time
}

type RetentionMap map[string]time.Time

func NewRetentionMap(ctx context.Context, retentionPeriod time.Duration) RetentionMap {
	rm := make(RetentionMap)

	go rm.Retention(ctx, retentionPeriod)

	return rm
}

func (m RetentionMap) NewOTP() OTP {
	otp := OTP{
		Key:     uuid.NewString(),
		Created: time.Now(),
	}
	m[otp.Key] = otp.Created.Add(5 * time.Minute)
	return otp
}

func (m RetentionMap) ValidateOTP(key string) bool {
	if expiry, ok := m[key]; ok {
		if time.Now().Before(expiry) {
			delete(m, key)
			return true
		}
		delete(m, key)
	}
	return false
}

func (m RetentionMap) Retention(ctx context.Context, retentionPeriod time.Duration) {
	ticker := time.NewTicker(400 * time.Millisecond)

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			for key, expiry := range m {
				if now.After(expiry) {
					delete(m, key)
				}
			}
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}
