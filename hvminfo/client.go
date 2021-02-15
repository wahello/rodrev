package hvminfo

import (
	"bufio"
	"fmt"
	"github.com/pkg/term"
	"go.uber.org/zap"
)

type ConfigClient struct {
	Port string `yaml:"port"`
	Speed int `yaml:"baudrate"`
	Logger *zap.SugaredLogger `yaml:"-"`
}



func RunClient (cfg *ConfigClient) error{
	if cfg == nil {return fmt.Errorf("no config passed")}
	if len(cfg.Port) < 1 {return fmt.Errorf("need Port parameter")}
	t, err := term.Open(cfg.Port, term.Speed(cfg.Speed), term.RawMode,term.FlowControl(term.NONE))
	if err != nil {return fmt.Errorf("error opening %s: %s",cfg.Port,err)}
	go func() {
		scanner := bufio.NewScanner(t)
		for scanner.Scan() {
			line := scanner.Text()
			cfg.Logger.Infof("got [%s] on serial\n",line)
		}
		cfg.Logger.Infof("serial reader exited\n")
	}()
	return nil
}