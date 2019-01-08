// Copyright 2018 Telefónica
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"io/ioutil"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	promcli "github.com/ghostbaby/prometheus-kafka-adapter/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/linkedin/goavro"
	"fmt"
	"net/http"
	"bytes"
	"strconv"
)

// Serializer represents an abstract metrics serializer
type Serializer interface {
	Marshal(metric map[string]interface{}) ([]byte, error)
}

type PodInfo struct {
	Pod_IP	string `json:"pod_ip"`
	Pod_Name	string	`json:"pod_name"`
}

//func GetPodIP(np string, name string) (error, string) {
//	config, err := rest.InClusterConfig()
//	if err != nil {
//		panic(err.Error())
//	}
//
//	// Creates the dynamic interface.
//	clientset, err := kubernetes.NewForConfig(config)
//	if err != nil {
//		panic(err.Error())
//	}
//
//	//podsListWatcher := cache1.NewListWatchFromClient(clientset.CoreV1().RESTClient(),"pod",np,fields.Everything())
//	//queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
//	pods, err := clientset.CoreV1().Pods(np).Get(name,metav1.GetOptions{})
//
//	if err != nil {
//		_ = pods
//		//log.Printf("ServiceMonitor %s is exists\n", sm.GetName())
//		return err,""
//	}
//
//	podIP := pods.Status.PodIP
//
//	return nil,podIP
//
//}

func GetPodIP(np string, name string,k8swatch string) (error, string) {
	res, err := http.Post(k8swatch,"",bytes.NewBuffer([]byte(name)))
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
	}

	defer res.Body.Close()

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
	}
	var podInfo PodInfo
	if err := json.Unmarshal(content, &podInfo); err == nil {
		return nil,podInfo.Pod_IP
	}else {
		return err,""

	}
}

// Serialize generates the JSON representation for a given Prometheus metric.
func Serialize(s Serializer, req *prompb.WriteRequest,k8swatch string, promeURL string) ([][]byte, error) {
	result := [][]byte{}

	for _, ts := range req.Timeseries {

		labels := make(map[string]string, len(ts.Labels))

		for _, l := range ts.Labels {
			labels[string(model.LabelName(l.Name))] = string(model.LabelValue(l.Value))
		}


		for _, sample := range ts.Samples {
			metricsName := string(labels["__name__"])
			metricsNamespace := string(labels["namespace"])
			metricsContainerName := string(labels["container_name"])
			//if strings.Contains(metricsName, "container") &&
			if metricsName == "container_cpu_usage_seconds_total" &&
				metricsNamespace == "dev" &&
				metricsContainerName != "POD"{
				//epoch := time.Unix(sample.Timestamp/1000, 0).Unix()
				metricsService := string(labels["service"])
				var endpoint string
				if metricsService == "kubelet" {
					endpoint = string(labels["pod_name"])
				}else if metricsService == "kube-state-metrics" {
					endpoint = string(labels["pod"])
				}

				err,podIP := GetPodIP(metricsNamespace,endpoint,k8swatch)
				if err != nil {
					fmt.Println(err)
					fmt.Println(labels)
					return nil,err
				}

				timestamp,value,err := promcli.GetPromContainerCpuUsage(endpoint,promeURL,sample.Timestamp)
				if err != nil {
					return nil,err
				}

				reqcpuName := string(labels["container_name"]) + "_req_cpu"
				fmt.Println(reqcpuName)

				err,reqCPU := GetPodIP(metricsNamespace,reqcpuName,k8swatch)
				fmt.Printf("pod %s request cpu is : %s",reqcpuName,reqCPU)
				if err != nil {
					return nil,err
				}

				reqCpuFlat64,err := strconv.ParseFloat(reqCPU,64)
				if err != nil {
					return nil,err
				}

				cpuPer := value / reqCpuFlat64


				m := map[string]interface{}{
					//"timestamp": epoch.Format(time.RFC3339),
					//"timestamp": time.Unix(sample.Timestamp/1000, 0).Unix(),
					"timestamp": timestamp,
					//"value":     strconv.FormatFloat(sample.Value, 'f', -1, 64),
					"value": cpuPer,
					"metric":      metricsName,
					"endpoint":	endpoint,
					"ip": podIP,
					"tags":    labels,
					"counterType": "GAUGE",
					"setp": 30,
				}

				data, err := s.Marshal(m)
				if err != nil {
					logrus.WithError(err).Errorln("couldn't marshal timeseries.")
				}

				result = append(result, data)
			}
		}
	}

	return result, nil
}

// JSONSerializer represents a metrics serializer that writes JSON
type JSONSerializer struct {
}

func (s *JSONSerializer) Marshal(metric map[string]interface{}) ([]byte, error) {
	return json.Marshal(metric)
}

func NewJSONSerializer() (*JSONSerializer, error) {
	return &JSONSerializer{}, nil
}

// AvroJSONSerializer represents a metrics serializer that writes Avro-JSON
type AvroJSONSerializer struct {
	codec *goavro.Codec
}

func (s *AvroJSONSerializer) Marshal(metric map[string]interface{}) ([]byte, error) {
	return s.codec.TextualFromNative(nil, metric)
}

// NewAvroJSONSerializer builds a new instance of the AvroJSONSerializer
func NewAvroJSONSerializer(schemaPath string) (*AvroJSONSerializer, error) {
	schema, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		logrus.WithError(err).Errorln("couldn't read avro schema")
		return nil, err
	}

	codec, err := goavro.NewCodec(string(schema))
	if err != nil {
		logrus.WithError(err).Errorln("couldn't create avro codec")
		return nil, err
	}

	return &AvroJSONSerializer{
		codec: codec,
	}, nil
}
