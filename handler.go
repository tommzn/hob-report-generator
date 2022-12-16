package main

import (
	"context"
	"time"

	"github.com/aws/aws-lambda-go/events"
	timetracker "github.com/tommzn/hob-timetracker"
)

// Run will fetch time tracking for last month, calculates a report, format it and distribute this report to a defined target.
func (handler *ReportGenerator) Run(event events.CloudWatchEvent, ctx context.Context) error {

	timeRangeStart, timeRangeEnd := reportTimeRange()
	handler.logger.Debugf("Generate report for %s - %s", timeRangeStart.Format("2006-01-02T15:04:05"), timeRangeEnd.Format("2006-01-02T15:04:05"))

	year := timeRangeStart.Year()
	month := int(timeRangeStart.Month())
	var timeTrackingRecords []timetracker.TimeTrackingRecord
	for _, deviceId := range handler.deviceIds {
		deviceRecords, err := handler.timeTracker.ListRecords(deviceId, timeRangeStart, timeRangeStart)
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

	reportFileName := timeRangeStart.Format("200601") + handler.formatter.FileExtension()
	return handler.publisher.Send(reportBuffer.Bytes(), reportFileName)
}

// ReportTimeRange generates first amd last day for report time range.
func reportTimeRange() (time.Time, time.Time) {
	now := time.Now()
	firstOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	return firstOfThisMonth.AddDate(0, -1, 0), firstOfThisMonth.Add(-1 * time.Second)
}
