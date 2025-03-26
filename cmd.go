package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "xquant",
	Short: "xquant server bootstrap cmd",
	Run:   serverBootstrap,
}

func serverBootstrap(cmd *cobra.Command, args []string) {
	//metrics.Init()

	go func() {
		runGrpc()
	}()

	runHTTP()
}
