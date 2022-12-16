package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/suite"
	timetracker "github.com/tommzn/hob-timetracker"
	"testing"
	"time"
)

type HandlerTestSuite struct {
	suite.Suite
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

func (suite *HandlerTestSuite) TestGenerateRepor() {

	handler := suite.handlerForTest()
	suite.Nil(handler.Run(events.CloudWatchEvent{}, context.Background()))
}

func (suite *HandlerTestSuite) handlerForTest() *ReportGenerator {

	conf := configForTest()
	locale := newLocale(conf)
	logger := loggerForTest()

	deviceIds := deviceIds(conf)
	formatter := newReportFormatter()
	calculator := newReportCalulator(locale)

	return &ReportGenerator{
		logger:      logger,
		deviceIds:   deviceIds,
		timeTracker: timeTrackeForTest(),
		calculator:  calculator,
		formatter:   formatter,
		publisher:   timetracker.NewFilePublisher(),
	}
}

func timeTrackeForTest() timetracker.TimeTracker {

	device := "Device01"
	tracker := timetracker.NewLocaLRepository()

	now := time.Now()
	firstOfThisMonth := time.Date(now.Year(), now.Month(), 1, 8, 0, 0, 0, now.Location())
	firstOfLastMonth := firstOfThisMonth.AddDate(0, -1, 0)
	tracker.Captured(device, timetracker.WORKDAY, firstOfLastMonth)
	tracker.Captured(device, timetracker.WORKDAY, firstOfLastMonth.Add(7*time.Hour))
	return tracker
}
