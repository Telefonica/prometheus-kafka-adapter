package prometheus

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

type Client struct {
	Server *url.URL
}


func onError(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func NewClient(addr string) (*Client, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	return &Client{
		Server: u,
	}, nil
}

type Qdata struct {
	ResultType string	`json:"result_type"`
	Result []*QueryRangeResponseResult     `json:"result"`
}

type Qreponse struct {
	Status string                  `json:"status"`
	Data   *Qdata `json:"data"`
}

type QueryRangeResponseResult struct {
	Metric map[string]string          `json:"metric"`
	Value *QueryRangeResponseValue `json:"value"`
}

type QueryRangeResponseValue []interface{}



func (c *Client) Query(query string) (*Qreponse, error) {
	u, err := url.Parse(fmt.Sprintf("./api/v1/query?query=%s",
		url.QueryEscape(query),
	))
	if err != nil {
		return nil, err
	}

	u = c.Server.ResolveReference(u)
	r, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	fmt.Println(string(b))

	if 400 <= r.StatusCode {
		return nil, fmt.Errorf("error response: %s", string(b))
	}

	resp := &Qreponse{}
	err = json.Unmarshal(b, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (v *QueryRangeResponseValue) Value() (float64, error) {
	s := (*v)[1].(string)
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}

	return f, nil
}

func (v *QueryRangeResponseValue) Time() time.Time {
	t := (*v)[0].(float64)
	return time.Unix(int64(t), 0)
}


func GetPromContainerCpuUsage(pod_name string,prom_url string,sample int64) (timestamp string,value string,err error){
	//query_str := "100 * (1 - avg by(instance_type, availability_zone)(irate(node_cpu{mode='idle'}[5m])))"
	query_str := "sum by (container_name) (rate(container_cpu_usage_seconds_total{job='kubelet', image!='',container_name!='POD',pod_name='" + pod_name + "'}[1m]))"

	client,err := NewClient(prom_url)
	if err != nil {
		onError(err)
	}
	resp,err := client.Query(query_str)
	if err != nil{
		onError(err)
	}
	var (
		tm string
		vl string
		vle float64
	)
	for _, r := range resp.Data.Result {
		vle, err = r.Value.Value()
		if err != nil {
			return "","",err
		}

		tm = strconv.FormatInt(r.Value.Time().Unix(),10)
	}
	vl = strconv.FormatFloat(vle,'f', -1, 64)
	return tm,vl,nil
	
}