package main

import (
	"fmt"

	"github.com/Cocoon-break/csync"
)

func main() {
	notifyCh := make(chan csync.NotifyData, 3)
	cs, err := csync.New(
		csync.WithComponent("demo"),
		csync.WithBasicAuth("password"),
		csync.WithTargetUrl("http://demo.com/api/v1/component-strategies"),
		csync.WithTagFunc(func() string { return "localhost" }),
		csync.WithNotifyCh(notifyCh),
		csync.WithIntervalSecond(60),
		csync.WithDumpPath("/tmp/strategies.json"),
	)
	if err != nil {
		fmt.Printf("unexpected error: %v", err)
		return
	}
	cs.Start()
	for c := range notifyCh {
		if c.Err != nil {
			fmt.Printf("unexpected error: %v\n", c.Err)
			continue
		}
		fmt.Printf("strategy map: %v\n", c.StrategyMap)
	}
}
