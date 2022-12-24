package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	core "github.com/tommzn/hob-core"
	timetracker "github.com/tommzn/hob-timetracker"
)

// HandleEvents will process given SQS events to generate time tracking reports.
func (handler *ReportGenerator) HandleEvents(ctx context.Context, sqsEvent events.SQSEvent) error {

	defer handler.logger.Flush()

	for _, message := range sqsEvent.Records {

		handler.logger.Debugf("Process message %s for event source %s", message.MessageId, message.EventSource)

		request := &core.GenerateReportRequest{}
		if err := core.DeserializeEvent(message.Body, request); err != nil {
			handler.logger.Error("Unable to deserialize event, reason: ", err)
			return err
		}

		formatter, err := newReportFormatter(request)
		if err != nil {
			handler.logger.Error("Unable to create formatter, reason: ", err)
			return err
		}
		handler.formatter = formatter

		publisher, err := newReportPublisher(handler.awsConf, request)
		if err != nil {
			handler.logger.Error("Unable to create publisher, reason: ", err)
			return err
		}
		handler.publisher = publisher

		if err := handler.GenerateReport(request); err != nil {
			handler.logger.Error("Unable to generate report, reason: ", err)
			return err
		}
	}
	return nil
}

// GenerateReport will generate a report based on passed type.
func (handler *ReportGenerator) GenerateReport(request *core.GenerateReportRequest) error {

	switch request.Type {

	case core.ReportType_MONTHLY_REPORT:
		return handler.GenerateMonthlyReport(request)

	default:
		err := fmt.Errorf("Unsupported report type: %s", request.Type)
		handler.logger.Error(err)
		return err
	}
}

// GenerateMonthlyReport will fetch time tracking for last month, calculates a report, format it and distribute this report to a defined target.
func (handler *ReportGenerator) GenerateMonthlyReport(request *core.GenerateReportRequest) error {

	timeRangeStart, timeRangeEnd := reportTimeRange(request)
	handler.logger.Debugf("Generate report for %s - %s", timeRangeStart.Format("2006-01-02T15:04:05"), timeRangeEnd.Format("2006-01-02T15:04:05"))

	year := timeRangeStart.Year()
	month := int(timeRangeStart.Month())
	var timeTrackingRecords []timetracker.TimeTrackingRecord
	for _, deviceId := range handler.deviceIds {
		deviceRecords, err := handler.timeTracker.ListRecords(deviceId, timeRangeStart, timeRangeEnd)
		if err != nil {
			return err
		}
		timeTrackingRecords = append(timeTrackingRecords, deviceRecords...)
	}
	handler.calculator.WithTimeTrackingRecords(timeTrackingRecords)

	if handler.calendar != nil {
		if holidays, err := handler.calendar.GetHolidays(year, month); err == nil {
			handler.formatter.WithHolidays(holidays)
		}
	}

	monthlyReport, err := handler.calculator.MonthlyReport(year, month, timetracker.WORKDAY)
	if err != nil {
		return err
	}

	reportBuffer, err := handler.formatter.WriteMonthlyReportToBuffer(monthlyReport)
	if err != nil {
		return err
	}

	reportFileName := timeRangeStart.Format(request.NamePattern) + handler.formatter.FileExtension()
	handler.logger.Debugf("Publish %s using %T", reportFileName, handler.publisher)
	return handler.publisher.Send(reportBuffer.Bytes(), reportFileName)
}

// ReportTimeRange generates first amd last day for report time range.
func reportTimeRange(request *core.GenerateReportRequest) (time.Time, time.Time) {

	if request.Year >= 2000 &&
		request.Month >= 1 && request.Month <= 12 {
		firstOfThisMonth := time.Date(int(request.Year), time.Month(request.Month), 1, 0, 0, 0, 0, time.UTC)
		return firstOfThisMonth, firstOfThisMonth.AddDate(0, 1, 0).Add(-1 * time.Second)
	} else {
		now := time.Now()
		firstOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		return firstOfThisMonth.AddDate(0, -1, 0), firstOfThisMonth.Add(-1 * time.Second)
	}
}

// NewReportFormatter returns an Excel report formatter with default settings.
func newReportFormatter(request *core.GenerateReportRequest) (timetracker.ReportFormatter, error) {

	switch request.Format {
	case core.ReportFormat_EXCEL:
		return timetracker.NewExcelReportFormatter(), nil
	default:
		return nil, fmt.Errorf("Unsupported report format: %s", request.Format)
	}
}

// NewReportPublisher returns a publisher to ditribute a report to a target defined in given report generate request.
func newReportPublisher(awsConf awsConfig, request *core.GenerateReportRequest) (timetracker.ReportPublisher, error) {

	if request.Delivery.S3 != nil {

		region := awsConf.region
		if request.Delivery.S3.Region != "" {
			region = &request.Delivery.S3.Region
		}

		bucket := awsConf.bucket
		if request.Delivery.S3.Bucket != "" {
			bucket = &request.Delivery.S3.Bucket
		}

		basePath := awsConf.basePath
		if request.Delivery.S3.Path != "" {
			basePath = &request.Delivery.S3.Path
		}
		return timetracker.NewS3Publisher(region, bucket, basePath), nil
	}

	if request.Delivery.File != nil {
		return timetracker.NewFilePublisher(&request.Delivery.File.Path), nil
	}

	return nil, errors.New("No report delivery defined!")
}
