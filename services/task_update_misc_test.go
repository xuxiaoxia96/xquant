package services

import (
	"testing"

	"xquant/models"
)

func TestRealtimeUpdateExchangeAndSnapshot(t *testing.T) {
	models.SyncAllSnapshots(nil)
	realtimeUpdateMiscAndSnapshot()
}
