package main

import (
	"errors"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"

	config "github.com/tommzn/go-config"
	log "github.com/tommzn/go-log"
	secrets "github.com/tommzn/go-secrets"
	timetracker "github.com/tommzn/hob-timetracker"
)

func main() {

	handler, err := bootstrap()
	if err != nil {
		panic(err)
	}
	lambda.Start(handler.Run)
}

// bootstrap loads config and other dependencies to creates a reprt generator.
func bootstrap() (*ReportGenerator, error) {

	conf, err := loadConfig()
	if err != nil {
		return nil, err
	}
	secretsManager := newSecretsManager()
	logger := newLogger(conf, secretsManager)

	awsConf, err := getAwsConfig(conf)
	if err != nil {
		return nil, err
	}

	timeTracker := newTimeTracker(awsConf)
	publisher := newReportPublisher(awsConf)
	deviceIds := deviceIds(conf)
	formatter := newReportFormatter()
	locale := newLocale(conf)
	calculator := newReportCalulator(locale)
	calendar, err := newCalendar(conf, secretsManager, locale)
	if err != nil {
		return nil, err
	}

	return &ReportGenerator{
		logger:      logger,
		deviceIds:   deviceIds,
		timeTracker: timeTracker,
		calculator:  calculator,
		formatter:   formatter,
		publisher:   publisher,
		calendar:    calendar,
	}, nil
}

// loadConfig from config file.
func loadConfig() (config.Config, error) {

	configSource, err := config.NewS3ConfigSourceFromEnv()
	if err != nil {
		return nil, err
	}
	return configSource.Load()
}

// newSecretsManager retruns a new secrets manager from passed config.
func newSecretsManager() secrets.SecretsManager {
	return secrets.NewSecretsManager()
}

// newLogger creates a new logger from  passed config.
func newLogger(conf config.Config, secretsMenager secrets.SecretsManager) log.Logger {
	logger := log.NewLoggerFromConfig(conf, secretsMenager)
	return log.WithNameSpace(logger, "hob-apigw-handler")
}

// NewTimeTracker creates a new time tracker to persist records in a S3 bucket.
func newTimeTracker(awsConf awsConfig) timetracker.TimeTracker {
	return timetracker.NewS3Repository(awsConf.region, awsConf.bucket, awsConf.basePath)
}

// GetAwsConfig extracts YWS region, bucket and a base path from given config.
func getAwsConfig(conf config.Config) (awsConfig, error) {
	region := conf.Get("aws.s3.region", config.AsStringPtr(os.Getenv("AWS_REGION")))
	bucket := conf.Get("aws.s3.bucket", nil)
	if bucket == nil {
		return awsConfig{}, errors.New("Np S3 bucket specified!")
	}
	basePath := conf.Get("aws.s3.basepath", nil)
	return awsConfig{region: region, bucket: bucket, basePath: basePath}, nil
}
func newReportPublisher(awsConf awsConfig) timetracker.ReportPublisher {
	return timetracker.NewS3Publisher(awsConf.region, awsConf.bucket, awsConf.basePath)
}

// NewLocale creates a new locale from given config.
func newLocale(conf config.Config) timetracker.Locale {

	country := conf.Get("hob.locale.country", config.AsStringPtr("de"))
	timezone := conf.Get("hob.locale.timezone", nil)
	dateformat := conf.Get("hob.locale.dateformat", nil)
	defailtWorktime := conf.GetAsDuration("hob.locale.defalt_worktime", config.AsDurationPtr(8*time.Hour))

	breaks := make(map[time.Duration]time.Duration)
	breakConfig := conf.GetAsSliceOfMaps("hob.locale.breaks")
	for _, breakConf := range breakConfig {
		worktimeStr, ok1 := breakConf["worktime"]
		breaktimeStr, ok2 := breakConf["breaktime"]
		if ok1 && ok2 {
			worktime := config.AsDuration(worktimeStr)
			breaktime := config.AsDuration(breaktimeStr)
			if worktime != nil && breaktime != nil {
				breaks[*worktime] = *breaktime
			}
		}
	}
	return timetracker.Locale{
		Country:         *country,
		Timezone:        timezone,
		DateFormat:      dateformat,
		DefaultWorkTime: *defailtWorktime,
		Breaks:          breaks,
	}
}

// NewReportCalulator will return a new report generator with given dependencies.
func newReportCalulator(location timetracker.Locale) timetracker.ReportCalculator {
	return timetracker.NewReportCalulator([]timetracker.TimeTrackingRecord{}, location)
}

// NewReportFormatter returns an Excel report formatter with default settings.
func newReportFormatter() timetracker.ReportFormatter {
	return timetracker.NewExcelReportFormatter()
}

// DeviceIds extracts a list device ids from passed config.
func deviceIds(conf config.Config) []string {
	ids := []string{}
	deviceConfig := conf.GetAsSliceOfMaps("hob.devices")
	for _, deviceConf := range deviceConfig {
		if id, ok := deviceConf["id"]; ok {
			ids = append(ids, id)
		}
	}
	return ids
}

// NewCalendar creates a new calendar api with given config and dependencies.
// An api key is required. Passed secrets manager uses key HOB_CALENDAR_APIKEY to obtain it.
func newCalendar(conf config.Config, secretsManager secrets.SecretsManager, location timetracker.Locale) (timetracker.Calendar, error) {
	apiKey, err := secretsManager.Obtain("HOB_CALENDAR_APIKEY")
	if err != nil {
		return nil, err
	}
	return timetracker.NewCalendarApi(*apiKey, location), nil
}
