package main

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/suite"
	core "github.com/tommzn/hob-core"
	timetracker "github.com/tommzn/hob-timetracker"
)

type HandlerTestSuite struct {
	suite.Suite
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

func (suite *HandlerTestSuite) TestGenerateReport() {

	handler := suite.handlerForTest()
	event := suite.sqsEventForTest(eventForTest())

	suite.Nil(handler.HandleEvents(context.Background(), event))

	event.Records[0].Body = "xxx"
	suite.NotNil(handler.HandleEvents(context.Background(), event))

	event2 := suite.sqsEventForTest(eventWithInvalidFormatterForTest())
	suite.NotNil(handler.HandleEvents(context.Background(), event2))

	event3 := suite.sqsEventForTest(eventWithoutDeliveryForTest())
	suite.NotNil(handler.HandleEvents(context.Background(), event3))

	event4 := suite.sqsEventForTest(eventWithInvalidTypeForTest())
	suite.NotNil(handler.HandleEvents(context.Background(), event4))
}

func (suite *HandlerTestSuite) TestGetReportTimeRange() {

	year := 2022
	month := 1
	start1, end1 := reportTimeRange(&core.GenerateReportRequest{Year: int64(year), Month: int64(month)})
	suite.Equal(1, start1.Day())
	suite.Equal(31, end1.Day())
	suite.Equal(month, int(start1.Month()))
	suite.Equal(month, int(end1.Month()))
	suite.Equal(year, start1.Year())
	suite.Equal(year, end1.Year())

	start2, end2 := reportTimeRange(&core.GenerateReportRequest{})
	suite.Equal(1, start2.Day())
	suite.True(end2.Day() >= 28)
}

func (suite *HandlerTestSuite) handlerForTest() *ReportGenerator {

	conf := configForTest()
	locale := newLocale(conf)
	logger := loggerForTest()

	deviceIds := deviceIds(conf)
	calculator := newReportCalulator(locale)

	return &ReportGenerator{
		awsConf:     awsConfig{},
		logger:      logger,
		deviceIds:   deviceIds,
		timeTracker: timeTrackeForTest(),
		calculator:  calculator,
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

func (suite *HandlerTestSuite) sqsEventForTest(event *core.GenerateReportRequest) events.SQSEvent {
	eventData, err := core.SerializeEvent(event)
	suite.Nil(err)
	return events.SQSEvent{
		Records: []events.SQSMessage{
			events.SQSMessage{
				MessageId:   "<ID>",
				EventSource: "<Source>",
				Body:        eventData,
			},
		},
	}
}

func eventForTest() *core.GenerateReportRequest {
	return &core.GenerateReportRequest{
		Format:      core.ReportFormat_EXCEL,
		Type:        core.ReportType_MONTHLY_REPORT,
		Year:        2022,
		Month:       1,
		NamePattern: "TestReport_200601",
		Delivery: &core.ReportDelivery{
			File: &core.FileTarget{
				Path: "./",
			},
		},
	}
}

func eventWithInvalidFormatterForTest() *core.GenerateReportRequest {
	return &core.GenerateReportRequest{
		Format:      core.ReportFormat_NO_FORMAT,
		Type:        core.ReportType_MONTHLY_REPORT,
		Year:        2022,
		Month:       1,
		NamePattern: "TestReport_200601",
		Delivery: &core.ReportDelivery{
			File: &core.FileTarget{
				Path: "./",
			},
		},
	}
}

func eventWithoutDeliveryForTest() *core.GenerateReportRequest {
	return &core.GenerateReportRequest{
		Format:      core.ReportFormat_EXCEL,
		Type:        core.ReportType_MONTHLY_REPORT,
		Year:        2022,
		Month:       1,
		NamePattern: "TestReport_200601",
		Delivery:    &core.ReportDelivery{},
	}
}

func eventWithInvalidTypeForTest() *core.GenerateReportRequest {
	return &core.GenerateReportRequest{
		Format:      core.ReportFormat_EXCEL,
		Type:        core.ReportType_NO_TYPE,
		Year:        2022,
		Month:       1,
		NamePattern: "TestReport_200601",
		Delivery: &core.ReportDelivery{
			File: &core.FileTarget{
				Path: "./",
			},
		},
	}
}
