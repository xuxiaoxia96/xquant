package trader

import (
	"context"
	"testing"
	"xquant/pkg/models"
)

func TestCacheSync(t *testing.T) {
	barIndex := 1
	models.SyncAllSnapshots(context.Background(), &barIndex)
	//UpdatePositions()
	SyncPositions()
	CacheSync()
}
