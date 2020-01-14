// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

const (
	maxBackoffSecs   = 300
	backoffResetSecs = 1800
)

type Gotifier struct {
	// Nick stores the nickname specified in the config, because irc.Client
	// might change its copy.
	Url            string
	StopRunning    chan bool
	StoppedRunning chan bool
	AlertNotices   chan AlertNotice
	BackoffCounter Delayer
}

func NewGotifier(config *Config, alertNotices chan AlertNotice) (*Gotifier, error) {
	backoffCounter := NewBackoff(
		maxBackoffSecs, backoffResetSecs,
		time.Second)

	notifier := &Gotifier{
		Url:            fmt.Sprintf("%s/message?token=%s", config.GotifyUrl, config.GotifyApiKey),
		StopRunning:    make(chan bool),
		StoppedRunning: make(chan bool),
		AlertNotices:   alertNotices,
		BackoffCounter: backoffCounter,
	}

	return notifier, nil
}

func (notifier *Gotifier) MaybeSendNotice(alertNotice *AlertNotice) {
	http.PostForm(notifier.Url,
		url.Values{"message": {alertNotice.Message}, "title": {fmt.Sprintf("%s Alert on %s", alertNotice.Status, alertNotice.Instance)}})
}

func (notifier *Gotifier) Run() {
	keepGoing := true
	for keepGoing {
		notifier.BackoffCounter.Delay()
		select {
		case alertNotice := <-notifier.AlertNotices:
			notifier.MaybeSendNotice(&alertNotice)
		case <-notifier.StopRunning:
			log.Printf("Gotify routine asked to terminate")
			keepGoing = false
		}
	}
	notifier.StoppedRunning <- true
}
