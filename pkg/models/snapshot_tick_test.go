package models

import (
	"context"
	"testing"
)

func TestSyncAllSnapshots(t *testing.T) {
	barIndex := 1
	SyncAllSnapshots(context.Background(), &barIndex)
}
