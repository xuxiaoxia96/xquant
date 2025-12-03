package main

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"

	"xquant/pkg/models"
)

var (
	onceApp     sync.Once
	application = ""
)

func lazyLoadApplication() {
	path, _ := os.Executable()
	_, exec := filepath.Split(path)
	application = exec
}

// ApplicationName 获取执行文件名
func ApplicationName() string {
	onceApp.Do(lazyLoadApplication)
	return application
}

var (
	Application = "stock"
	MinVersion  = "0.0.1" // 主程序版本号
)

func init() {
	Application = ApplicationName()
}

var rootCmd = &cobra.Command{
	Use:   "xquant",
	Short: "xquant server bootstrap cmd",
	Run:   serverBootstrap,
}

var (
	SnapshotManager *models.SnapshotManager
)

func serverBootstrap(cmd *cobra.Command, args []string) {
	//metrics.Init()
	SnapshotManager = models.NewSnapshotManager()

	runHTTP()

}
