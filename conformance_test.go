package main

import (
	"fmt"
	"testing"

	"github.com/bubustack/bubu-sdk-go/conformance"
	"github.com/bubustack/json-filter-engram/pkg/config"
	"github.com/bubustack/json-filter-engram/pkg/engram"
)

func TestConformance(t *testing.T) {
	suite := conformance.BatchSuite[config.Config, config.Inputs]{
		Engram:      engram.New(),
		Config:      config.Config{},
		Inputs:      config.Inputs{},
		ExpectError: true,
		ValidateError: func(err error) error {
			if err == nil {
				return nil
			}
			if err.Error() != "input is required" {
				return fmt.Errorf("unexpected conformance error: %w", err)
			}
			return nil
		},
	}
	suite.Run(t)
}
