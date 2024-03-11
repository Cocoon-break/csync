# strategy sync sdk

sync the strategy content from center by http request 

## install
```
go get github.com/Cocoon-break/csync
```

## usage
```go
    notifyCh := make(chan NotifyData)
		cs, err := New(
			WithComponentName("component"),
			WithBasicAuth("user", "password"),
			WithTargetUrl("http://example.com"),
			WithTagFunc(func() string { return "tag" }),
			WithNotifyCh(notifyCh),
		)
		if err != nil {
			fmt.Printf("unexpected error: %v", err)
			return
		}
		cs.Start()
		go func() {
			for notifyData := range notifyCh {
				if notifyData.Err != nil {
					continue
				}
				// Add assertions for the received notify data
			}
		}()
		// application exit stop chan
		defer cs.Stop()
```