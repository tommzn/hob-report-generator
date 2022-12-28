package main

import (
	config "github.com/tommzn/go-config"
	log "github.com/tommzn/go-log"
	timetracker "github.com/tommzn/hob-timetracker"
)

// ReportGenerator will fetch time tracking records and generates reports.
type ReportGenerator struct {
	logger      log.Logger
	conf        config.Config
	awsConf     awsConfig
	deviceIds   []string
	timeTracker timetracker.TimeTracker
	calculator  timetracker.ReportCalculator
	formatter   timetracker.ReportFormatter
	publisher   []timetracker.ReportPublisher
	calendar    timetracker.Calendar
}

// AwsConfig used for different AWS clients.
type awsConfig struct {
	region, bucket, basePath *string
}

type awsEventBridgeTrigger struct {
	Content string `json:"content"`
}
