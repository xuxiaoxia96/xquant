package conf

import (
	"bytes"
	"embed"

	"github.com/spf13/viper"
)

//go:embed *.yaml
var embedConf embed.FS

func Load(confFileName string, confValue any) error {
	viper.SetConfigType("yaml")

	confFileBytes, err := embedConf.ReadFile(confFileName)
	if err != nil {
		return err
	}

	if err = viper.ReadConfig(bytes.NewReader(confFileBytes)); err != nil {
		return err
	}

	if err = viper.Unmarshal(confValue); err != nil {
		return err
	}

	return nil
}
