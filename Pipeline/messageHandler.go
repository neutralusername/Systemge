package Pipeline

import (
	"errors"

	"github.com/neutralusername/Systemge/Tools"
)

type Pipeline struct {
	rateLimiterBytes *Tools.TokenBucketRateLimiter
	rateLimiterCalls *Tools.TokenBucketRateLimiter
	deserializer     func([]byte) (any, error)
	validator        func(any) error
}

func (pipeline *Pipeline) DoStuff(bytes []byte) (any, error) {
	if pipeline.rateLimiterBytes != nil && !pipeline.rateLimiterBytes.Consume(uint64(len(bytes))) {
		return nil, errors.New("byte rate limit exceeded")
	}
	if pipeline.rateLimiterCalls != nil && !pipeline.rateLimiterCalls.Consume(1) {
		return nil, errors.New("call rate limit exceeded")
	}
	data, err := pipeline.deserializer(bytes)
	if err != nil {
		return nil, err
	}
	if pipeline.validator != nil {
		err = pipeline.validator(data)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}
