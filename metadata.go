package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
)

func GetAllMetricAttributes(promAPIEndPoint string, metricsList map[string]MetricAttributes) {

	// Make a GET request to the Prometheus metadata API
	response, err := http.Get(promAPIEndPoint)
	if err != nil {
		logrus.WithError(err).Errorln("Error making request")
		return
	}
	defer response.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logrus.WithError(err).Errorln("Error reading response body")
		// logrus.error("Error reading response body: %s\n", err.Error())
		return
	}

	// Parse the JSON data into a map
	var data map[string]interface{}
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		logrus.WithError(err).Errorln("Error parsing json")
		return
	}
	logrus.WithFields(logrus.Fields{
		"[]": len(data["data"].(map[string]interface{})),
	}).Debug("Metrics Count is ")
	// var metricList = make(map[string]MetricAttributes)
	for key, metrics := range data["data"].(map[string]interface{}) {
		for _, metric := range metrics.([]interface{}) {
			var metricAttribute MetricAttributes
			metricAttribute.metricType = metric.(map[string]interface{})["type"].(string)
			metricAttribute.metricHelp = metric.(map[string]interface{})["help"].(string)
			metricAttribute.metricUnit = metric.(map[string]interface{})["unit"].(string)
			metricsList[key] = metricAttribute
			// fmt.Printf("Metric: %s, Type: %s, Help: %s, Unit: %s", key, metricAttribute.metricType, metricAttribute.metricHelp, metricAttribute.metricUnit)
		}
	}
	logrus.WithFields(logrus.Fields{
		"[]": len(metricsList),
	}).Debug("Map Size is ")
}
