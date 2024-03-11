package csync

type GetTagFunc func() string

type Component string
type StrategyName string
type StrategyDetail struct {
	Id         int    `json:"id"`
	Alias      string `json:"alias"`
	Content    []byte `json:"content"`
	ContentMd5 string `json:"content_md5"`
}

// notify data if err not nil, means sync failed
// you should check Err is nil or not
type NotifyData struct {
	Err         error
	StrategyMap map[StrategyName]StrategyDetail
}

// request center server to sync strategy body
type syncReq struct {
	Hostname       string                  `json:"hostname"`
	ComponentName  Component               `json:"component_name"`
	StrategyMd5Map map[StrategyName]string `json:"strategy_md5_map"`
}

// center server to sync strategy response body
type syncResp struct {
	Code int                             `json:"code"`
	Msg  string                          `json:"msg"`
	Data map[StrategyName]StrategyDetail `json:"data"`
}
