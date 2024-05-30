package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

func GetAllMetricMetadata(promMetaDataEndPoint string, metricsList map[string]MetricMetadata) {

	// Make a GET request to the Prometheus metadata API
	response, err := http.Get(promMetaDataEndPoint)
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
	// var metricList = make(map[string]MetricMetadata)
	for key, metrics := range data["data"].(map[string]interface{}) {
		for _, metric := range metrics.([]interface{}) {
			var metricMetadata MetricMetadata
			logrus.Debugf("Processing Metric %s, Metadata to be included %s", key, includedMetaData)

			if strings.Contains(strings.ToLower(includedMetaData), "type") {
				metricMetadata.metricType = metric.(map[string]interface{})["type"].(string)
				logrus.Debugf("Type is %s", metricMetadata.metricType)
			}
			if strings.Contains(strings.ToLower(includedMetaData), "help") {
				metricMetadata.metricHelp = metric.(map[string]interface{})["help"].(string)
				logrus.Debugf("Help is %s", metricMetadata.metricHelp)
			}
			if strings.Contains(strings.ToLower(includedMetaData), "unit") {
				metricMetadata.metricUnit = metric.(map[string]interface{})["unit"].(string)
				logrus.Debugf("Unit is %s", metricMetadata.metricUnit)
			}
			metricsList[key] = metricMetadata
			// fmt.Printf("Metric: %s, Type: %s, Help: %s, Unit: %s", key, metricMetadata.metricType, metricMetadata.metricHelp, metricMetadata.metricUnit)
		}
	}
	logrus.Debugf("Total number of metrics parsed is %v", len(metricsList))
}
