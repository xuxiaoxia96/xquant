package main

import (
	"github.com/spf13/cobra"
	"xquant/pkg/models"
)

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

	go func() {
		runGrpc()
	}()

	SnapshotManager = models.NewSnapshotManager()

	runHTTP()

}
