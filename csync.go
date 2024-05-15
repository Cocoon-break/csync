package csync

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"
)

type Syncer interface {
	Start()
	Stop()
}
type sync struct {
	c           *config
	stop        chan struct{}
	strategyMap map[StrategyName]StrategyDetail
}

func New(opts ...Option) (Syncer, error) {
	c := newDefaultConfig()
	for _, opt := range opts {
		if err := opt.Apply(c); err != nil {
			return nil, err
		}
	}
	if err := c.validate(); err != nil {
		return nil, err
	}
	return &sync{
		c:           c,
		stop:        make(chan struct{}),
		strategyMap: make(map[StrategyName]StrategyDetail),
	}, nil
}

func (s *sync) Start() {
	// load last dump file
	tmpStrategyMap := loadFromDisk(s.c.DumpPath)
	if tmpStrategyMap != nil {
		s.notify(NotifyData{
			StrategyMap: tmpStrategyMap,
		})
		s.strategyMap = tmpStrategyMap
	}
	// data
	s.runTicker(s.c.IntervalSecond, s.syncStrategy)
}

func (s *sync) Stop() {
	s.stop <- struct{}{}
}

func (s *sync) syncStrategy() {
	strategyMd5Map := make(map[StrategyName]string)
	for k, v := range s.strategyMap {
		strategyMd5Map[k] = v.ContentMd5
	}
	body := syncReq{
		Endpoint:       s.c.GetTagFunc(),
		ComponentName:  s.c.Component,
		StrategyMd5Map: strategyMd5Map,
	}
	data, _ := json.Marshal(body)
	notifyData := s.sendRequest(data)
	if notifyData.Err == nil {
		avaMap := make(map[StrategyName]StrategyDetail, len(notifyData.StrategyMap))
		for k, v := range notifyData.StrategyMap {
			// find no changed data,fill last data to avaMap
			if d, ok := s.strategyMap[k]; ok && v.Content == _configNoChanged {
				avaMap[k] = d
			}
			// check md5,if equal,fill to avaMap
			if v.ContentMd5 == createStringMD5(v.Content) {
				avaMap[k] = v
			}
		}
		notifyData.StrategyMap = avaMap
		s.strategyMap = avaMap
		// ignore error
		dumpToDisk(s.c.DumpPath, s.strategyMap)
	}
	s.notify(notifyData)
}

func (s *sync) notify(data NotifyData) {
	if s.c.NotifyCh != nil {
		s.c.NotifyCh <- data
	}
}

func (s *sync) runTicker(interval int, fn func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				buf := make([]byte, 64<<10)
				buf = buf[:runtime.Stack(buf, false)]
				os.Stdout.Write(buf)
			}
		}()
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				fn()
			case <-s.stop:
				return
			}
		}
	}()
}

func (s *sync) sendRequest(reqBody []byte) (result NotifyData) {
	url := s.c.TargetUrl
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		result.Err = err
		return
	}
	// set auth and headers
	req.SetBasicAuth(string(s.c.Component), s.c.Password)
	req.Header.Set("Component", string(s.c.Component))

	cli := http.Client{Timeout: 30 * time.Second}
	resp, err := cli.Do(req)
	if err != nil {
		result.Err = err
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		result.Err = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		return
	}

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Err = err
		return
	}
	var respBody syncResp
	err = json.Unmarshal(respData, &respBody)
	if err != nil {
		result.Err = err
		return
	}
	result.StrategyMap = respBody.Data
	return
}

func loadFromDisk(path string) map[StrategyName]StrategyDetail {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var strategyMap map[StrategyName]StrategyDetail
	if err = json.Unmarshal(data, &strategyMap); err != nil {
		return nil
	}
	return strategyMap
}

func dumpToDisk(path string, data map[StrategyName]StrategyDetail) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, os.ModePerm)
}

func createStringMD5(message string) string {
	h := md5.New()
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}
