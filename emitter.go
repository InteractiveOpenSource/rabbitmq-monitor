package main

import (
	"sync"
	"fmt"
	"time"
	"encoding/json"
	event "github.com/wauio/event-emitter"
)

// singleton implementation
var emitter *event.EventEmitter
var once sync.Once

func GetEmitter() *event.EventEmitter {
	once.Do(func() {
		emitter = event.New()
	})

	return emitter
}

// usage and logic
func defineEmitter(e *event.EventEmitter) {
	e.On("log", func(scope string, level string, message string, data interface{}) error {
		var byteData []byte

		if val, isByte := data.([]byte); isByte {
			byteData = val
		} else {
			if byteData, err = json.Marshal(data); err != nil {
				byteData = []byte("[]")
			}
		}

		fmt.Printf("[%s] %s.%s: %s %s\n", time.Now().Format("2006-01-02 15:04:05"), scope, level, message, byteData)
		return nil
	})

	e.On("queues.data", func(data interface{}) error {
		var byteData []byte
		if byteData, err = json.Marshal(data); err != nil {
			byteData = []byte("[]")
		}

		fmt.Sprintf("%s", byteData)

		return nil
	})
}
