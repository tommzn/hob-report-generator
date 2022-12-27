[![Actions Status](https://github.com/tommzn/hob-report-generator/actions/workflows/go.image.build.yml/badge.svg)](https://github.com/tommzn/hob-report-generator/actions)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/tommzn/hob-report-generator)

# HomeOffice Button - Report Generator
Generates time tracking reports - monthly reports only atm. This report generator fetches time tracking records for a single year/month, generates a report in a defined format and distributes this report to a target.  
This report generator belongs to the [HomeOffice Button - Time Tracking](https://github.com/tommzn/hob-timetracker) Project. This repository contains all report formats, a report calculator and different target a report can be send to.

## Trigger
This report generator listen to a queue for a GenerateReportRequest. This request can provide a year and month a report should be generated for and defines report format and a target a report should be delivered to. As describes in time tracking project there're two types of trigger atm.
### Monthly Trigger
A rule defined in AWS EventBridger, e.g. at each 1st of a month, will publish an event to used SQS queue to trigger a report generation for previous month.
### API Gateway
External clients, e.g. an App, can trigger report generation via API.

# Links
[HomeOffice Button - Time Tracking](https://github.com/tommzn/hob-timetracker)  
[AWS IoT 1-Click](https://aws.amazon.com/iot-1-click/?nc1=h_ls)  
