package main

import (
	"github.com/stretchr/testify/suite"
	config "github.com/tommzn/go-config"
	log "github.com/tommzn/go-log"
	secrets "github.com/tommzn/go-secrets"
	"os"
	"testing"
	"time"
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

func (suite *BootstrapTestSuite) TestGetDeviceIds() {

	ids := deviceIds(configForTest())
	suite.Len(ids, 3)
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
