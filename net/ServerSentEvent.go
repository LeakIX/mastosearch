package net

import (
	"bufio"
	"bytes"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type ServerSentEvent struct {
	Event string
	Data  []byte
}

type ServerSentEventReader struct {
	url           string
	httpScanner   *bufio.Scanner
	outputChannel chan ServerSentEvent
}

var ErrConnectionFailed = errors.New("connection failed")
var ErrHelloFailed = errors.New("hello failed")

func NewServerSentEventReader(url string, outputChannel chan ServerSentEvent) (*ServerSentEventReader, error) {
	sser := &ServerSentEventReader{
		url:           url,
		outputChannel: outputChannel,
	}
	go sser.start()
	return sser, nil
}
func (sser *ServerSentEventReader) start() {
	retryTime := 30
	for {
		if err := sser.connect(); err != nil {
			log.Println(err)
		} else {
			retryTime = 30
			if err := sser.stream(); err != nil {
				log.Println(err)
			}
		}
		rand.Seed(time.Now().UnixNano())
		sleepTime := retryTime + rand.Intn(retryTime)
		if retryTime < 600 {
			retryTime += 30
		}
		log.Printf("reconnecting to %s in %d seconds", sser.url, sleepTime)
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}
}
func (sser *ServerSentEventReader) connect() error {
	resp, err := http.DefaultClient.Get(sser.url)
	if err != nil || resp.StatusCode != 200 {
		return ErrConnectionFailed
	}
	sser.httpScanner = bufio.NewScanner(resp.Body)
	// Use a 64KB buffer
	sser.httpScanner.Buffer(make([]byte, bufio.MaxScanTokenSize), bufio.MaxScanTokenSize)
	if !sser.httpScanner.Scan() {
		return ErrHelloFailed
	}
	if sser.httpScanner.Text() != ":)" {
		return ErrHelloFailed
	}
	return nil
}

func (sser *ServerSentEventReader) stream() error {
	for sser.httpScanner.Scan() {
		// Skip thumps
		if sser.httpScanner.Text() == ":thump" {
			continue
		}
		// Read event line
		eventLine := sser.httpScanner.Text()
		if !strings.HasPrefix(eventLine, "event: ") {
			return errors.New("scanner failed, server is now dead")
		}
		event := ServerSentEvent{
			Event: strings.Replace(eventLine, "event: ", "", 1),
		}
		// Read data line
		if !sser.httpScanner.Scan() {
			return errors.New("scanner failed, server is now dead")
		}
		event.Data = bytes.TrimLeft(sser.httpScanner.Bytes(), "data: ")
		// send our event
		sser.outputChannel <- event
		// Expect empty line
		if !sser.httpScanner.Scan() || sser.httpScanner.Text() != "" {
			return errors.New("scanner failed, server is now dead")
		}
	}
	return errors.New("end of stream")
}
