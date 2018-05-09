package metrics

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	yaml "gopkg.in/yaml.v2"
)

type UaaResp struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	Jti          string `json:"jti"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

type FoundationAppUsage struct {
	ReportTime     string `json:"report_time"`
	MonthlyReports []struct {
		Month               int     `json:"month"`
		Year                int     `json:"year"`
		AverageAppInstances float64 `json:"average_app_instances"`
		MaximumAppInstances int     `json:"maximum_app_instances"`
		AppInstanceHours    float64 `json:"app_instance_hours"`
	} `json:"monthly_reports"`
	YearlyReports []struct {
		Year                int     `json:"year"`
		AverageAppInstances float64 `json:"average_app_instances"`
		MaximumAppInstances int     `json:"maximum_app_instances"`
		AppInstanceHours    float64 `json:"app_instance_hours"`
	} `json:"yearly_reports"`
}

type Foundation struct {
	Name     string             `yaml:"name"`
	URL      string             `yaml:"url"`
	Password string             `yaml:"admin_password,omitempty"`
	AppUsage FoundationAppUsage `yaml:"omitempty"`
}

type Config struct {
	Foundations []Foundation `yaml:"foundations"`
}

func GetAppData(c *gin.Context) {

	var wg sync.WaitGroup
	var FoundationsAppUsage []Foundation
	config := getConfig()
	foundationsAppUsageChannel := make(chan Foundation, len(config.Foundations))

	for _, foundation := range config.Foundations {
		wg.Add(1)
		go getFoundationAppUsage(foundation, foundationsAppUsageChannel, &wg)
	}
	wg.Wait()
	close(foundationsAppUsageChannel)
	for foundationAppUsage := range foundationsAppUsageChannel {
		FoundationsAppUsage = append(FoundationsAppUsage, foundationAppUsage)
	}

	c.JSON(http.StatusOK, &FoundationsAppUsage)
}

func getConfig() Config {
	data, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	config := Config{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	if len(config.Foundations) == 0 {
		log.Fatalf("Something is wrong with your config.yml!")
	}
	if config.Foundations[0].URL == "" {
		log.Fatalf("The config.yml is missing a foundation URL!")
	}
	if config.Foundations[0].Password == "" {
		log.Fatalf("The config.yml is missing a foundation admin password!")
	}
	return config

}

func getToken(foundation Foundation) string {
	var username = "cf"
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("POST", "https://login."+foundation.URL+"/oauth/token?grant_type=password&password="+foundation.Password+"&username=admin", nil)
	req.SetBasicAuth(username, "")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	uaaResp := UaaResp{}
	if err := json.Unmarshal(body, &uaaResp); err != nil {
		log.Fatal(err)
	}
	return uaaResp.AccessToken
}

func getFoundationAppUsage(foundation Foundation, c chan Foundation, wg *sync.WaitGroup) {
	token := getToken(foundation)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", "https://app-usage."+foundation.URL+"/system_report/app_usages", nil)
	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	usageResp := FoundationAppUsage{}
	if err := json.Unmarshal(body, &usageResp); err != nil {
		log.Fatal(err)
	}
	foundationAppUsage := Foundation{}
	foundationAppUsage.Name = foundation.Name
	foundationAppUsage.URL = foundation.URL
	foundationAppUsage.AppUsage = usageResp
	c <- foundationAppUsage
	wg.Done()
}
