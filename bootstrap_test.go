package main

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	config "github.com/tommzn/go-config"
	log "github.com/tommzn/go-log"
	secrets "github.com/tommzn/go-secrets"
	core "github.com/tommzn/hob-core"
	timetracker "github.com/tommzn/hob-timetracker"
)

type BootstrapTestSuite struct {
	suite.Suite
}

func TestBootstrapTestSuite(t *testing.T) {
	suite.Run(t, new(BootstrapTestSuite))
}

func (suite *BootstrapTestSuite) TestCreateLocaleFromConfig() {

	conf := configForTest()
	locale := newLocale(conf)
	suite.Equal("NL", locale.Country)
	suite.NotNil(locale.Timezone)
	suite.Equal("Europe/Rom", *locale.Timezone)
	suite.NotNil(locale.DateFormat)
	suite.Equal("2006/01/02", *locale.DateFormat)

	suite.Len(locale.Breaks, 2)
	breaktime1, ok1 := locale.Breaks[6*time.Hour]
	suite.True(ok1)
	suite.Equal(30*time.Minute, breaktime1)
	breaktime2, ok2 := locale.Breaks[9*time.Hour]
	suite.True(ok2)
	suite.Equal(15*time.Minute, breaktime2)
}

func (suite *BootstrapTestSuite) TestGetAwsConfig() {

	conf := configForTest()
	awsConf, err := getAwsConfig(conf)
	suite.Nil(err)
	suite.NotNil(awsConf.region)
	suite.Equal("eu-central-5", *awsConf.region)
	suite.NotNil(awsConf.bucket)
	suite.Equal("test", *awsConf.bucket)
	suite.NotNil(awsConf.basePath)
	suite.Equal("path", *awsConf.basePath)

	conf2 := emptyConfigForTest()
	awsConf2, err2 := getAwsConfig(conf2)
	suite.NotNil(err2)
	suite.Nil(awsConf2.region)
	suite.Nil(awsConf2.bucket)
	suite.Nil(awsConf2.basePath)
}

func (suite *BootstrapTestSuite) TestCreateCalendar() {

	conf := configForTest()
	locale := newLocale(conf)

	_, err := newCalendar(conf, secretsManagerForTest(), locale)
	suite.NotNil(err)

	os.Setenv("HOB_CALENDAR_APIKEY", "xxx")
	_, err2 := newCalendar(conf, secretsManagerForTest(), locale)
	suite.Nil(err2)
}

func (suite *BootstrapTestSuite) TestNewReportFormatter() {

	formatter1, err1 := newReportFormatter(&core.GenerateReportRequest{Format: core.ReportFormat_EXCEL}, loggerForTest())
	suite.NotNil(formatter1)
	suite.Nil(err1)

	formatter2, err2 := newReportFormatter(&core.GenerateReportRequest{Format: core.ReportFormat_NO_FORMAT}, loggerForTest())
	suite.Nil(formatter2)
	suite.NotNil(err2)
}

func (suite *BootstrapTestSuite) TestNewReportPublisher() {

	handler := &ReportGenerator{awsConf: awsConfig{}, conf: configForTest(), logger: loggerForTest()}

	request1 := &core.GenerateReportRequest{
		Delivery: &core.ReportDelivery{
			S3: &core.S3Target{
				Region: "eu-central-5",
				Bucket: "bucket",
				Path:   "/base_path/",
			},
		},
	}
	publisher1, err1 := handler.newReportPublisher(request1)
	suite.NotNil(publisher1)
	suite.Nil(err1)

	request1_1 := &core.GenerateReportRequest{
		Delivery: &core.ReportDelivery{
			S3: &core.S3Target{},
		},
	}
	handler.awsConf = awsConfig{
		region:   asStringPtr("eu-central-5"),
		bucket:   asStringPtr("bucket"),
		basePath: asStringPtr("/base_path/"),
	}
	publisher1_1, err1_1 := handler.newReportPublisher(request1_1)
	suite.NotNil(publisher1_1)
	suite.Nil(err1_1)

	request2 := &core.GenerateReportRequest{
		Delivery: &core.ReportDelivery{
			File: &core.FileTarget{
				Path: "/tmp(",
			},
		},
	}
	publisher2, err2 := handler.newReportPublisher(request2)
	suite.NotNil(publisher2)
	suite.Nil(err2)

	request3 := &core.GenerateReportRequest{
		Delivery: &core.ReportDelivery{},
	}
	publisher3, err3 := handler.newReportPublisher(request3)
	suite.Len(publisher3, 0)
	suite.NotNil(err3)

	request4 := &core.GenerateReportRequest{
		Delivery: &core.ReportDelivery{
			Mail: &core.MailTarget{
				ToAddresses: []string{"user@example.com"},
			},
		},
	}
	publisher4, err4 := handler.newReportPublisher(request4)
	suite.NotNil(publisher4)
	suite.Len(publisher4, 1)
	suite.IsType(&timetracker.EMailPublisher{}, publisher4[0])
	suite.Nil(err4)
	suite.NotNil(publisher4[0].(*timetracker.EMailPublisher).Source)
	suite.True(len(publisher4[0].(*timetracker.EMailPublisher).Source) > 0)
}

func configForTest() config.Config {
	configFile := "fixtures/testconfig.yml"
	configLoader := config.NewFileConfigSource(&configFile)
	config, _ := configLoader.Load()
	return config
}

func emptyConfigForTest() config.Config {
	configFile := "fixtures/testconfig-empty.yml"
	configLoader := config.NewFileConfigSource(&configFile)
	config, _ := configLoader.Load()
	return config
}

// loggerForTest creates a new stdout logger for testing.
func loggerForTest() log.Logger {
	return log.NewLogger(log.Debug, nil, nil)
}

func secretsManagerForTest() secrets.SecretsManager {
	return &secrets.EnvironmentSecretsManager{}
}

func asStringPtr(s string) *string {
	return &s
}
