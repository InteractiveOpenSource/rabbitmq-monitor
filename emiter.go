package main

import (
	"sync"
	"encoding/json"
	"fmt"
	"time"
)

// Event emitter
type EventEmitter struct {
	listeners map[string][]Listener
}

func (e *EventEmitter) On(event string, listener Listener) {
	e.listeners[event] = append(e.listeners[event], listener)
}

func (e *EventEmitter) Fire(event string, anything interface{}) (response []interface{}) {
	payload := json.RawMessage{}
	payload, _ = json.Marshal(anything)

	return e.FireRaw(event, payload)
}

func (e *EventEmitter) FireRaw(event string, payload []byte) (response []interface{}) {
	for _, listener := range e.listeners[event] {
		response = append(response, listener(payload))
	}

	return
}

// Listener
type Listener func(payload []byte) interface{}

// singleton implementation
var emitter *EventEmitter
var once sync.Once

func GetEmitter() *EventEmitter {
	once.Do(func() {
		emitter = &EventEmitter{}

		// init listeners map
		emitter.listeners = make(map[string][]Listener)
	})

	return emitter
}

// usage and logic
func defineEmitter(e *EventEmitter) {
	e.On("log", func(payload []byte) interface {} {
		fmt.Printf("[%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), string(payload[:]))

		return nil
	})
}