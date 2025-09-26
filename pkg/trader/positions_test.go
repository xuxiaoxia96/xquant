package trader

import (
	"testing"
	"xquant/pkg/models"
)

func TestCacheSync(t *testing.T) {
	barIndex := 1
	models.SyncAllSnapshots(&barIndex)
	//UpdatePositions()
	SyncPositions()
	CacheSync()
}
