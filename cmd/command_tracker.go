package cmd

import (
	"context"
	"fmt"
	"strings"

	"gitee.com/quant1x/gox/api"
	cmder "github.com/spf13/cobra"

	trackerservice "xquant/biz/service/tracker"
	"xquant/pkg/log"
	"xquant/pkg/models"
)

const (
	trackerCommand     = "tracker"
	trackerDescription = "实时跟踪"
)

var (
	trackerStrategyCodes                = "1" // 策略编号
	Tracker              *cmder.Command = nil // 实时跟踪
)

func initTracker() {
	Tracker = &cmder.Command{
		Use:     trackerCommand,
		Example: cfg.Application + " " + trackerCommand + " --no=1",
		Args: func(cmd *cmder.Command, args []string) error {
			return nil
		},
		Short: trackerDescription,
		Long:  trackerDescription,
		Run: func(cmd *cmder.Command, args []string) {
			var strategyCodes []uint64
			array := strings.Split(trackerStrategyCodes, ",")
			for _, strategyNumber := range array {
				code := api.ParseUint(strings.TrimSpace(strategyNumber))
				_, err := models.CheckoutStrategy(code)
				if err != nil {
					fmt.Printf("策略编号%d, 不存在\n", code)
					log.Errorf("策略编号%d, 不存在", code)
					continue
				}

				strategyCodes = append(strategyCodes, code)
			}
			if len(strategyCodes) == 0 {
				fmt.Println("没有有效的策略编号, 实时扫描结束")
				log.Infof("没有有效的策略编号, 实时扫描结束")
				return
			}
			trackerservice.RunTrackerCore(context.Background(), trackerservice.TrackerCoreParams{TrackerStrategyCodes: strategyCodes, IsDebug: true})
		},
	}

	Tracker.Flags().StringVar(&trackerStrategyCodes, "no", trackerStrategyCodes, "策略编号, 多个用逗号分隔")
}
