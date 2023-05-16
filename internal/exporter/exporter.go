package exporter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	score = prometheus.NewDesc(
		prometheus.BuildFQName("awair", "", "score"),
		"Awair Score (0-100)",
		[]string{
			"device_uuid",
		},
		nil,
	)

	dew_point = prometheus.NewDesc(
		prometheus.BuildFQName("awair", "", "dew_point"),
		"The temperature at which water will condense and form into dew (ºC)",
		[]string{
			"device_uuid",
		},
		nil,
	)

	temp = prometheus.NewDesc(
		prometheus.BuildFQName("awair", "", "temp"),
		"Dry bulb temperature (ºC)",
		[]string{
			"device_uuid",
		},
		nil,
	)

	humidity = prometheus.NewDesc(
		prometheus.BuildFQName("awair", "", "humidity"),
		"Relative Humidity (%)",
		[]string{
			"device_uuid",
		},
		nil,
	)

	abs_humidity = prometheus.NewDesc(
		prometheus.BuildFQName("awair", "", "absolute_humidity"),
		"Absolute Humidity (g/m³)",
		[]string{
			"device_uuid",
		},
		nil,
	)

	co2 = prometheus.NewDesc(
		prometheus.BuildFQName("awair", "", "co2"),
		"Carbon Dioxide (ppm)",
		[]string{
			"device_uuid",
		},
		nil,
	)

	co2_estimated = prometheus.NewDesc(
		prometheus.BuildFQName("awair", "", "co2_est"),
		"Estimated Carbon Dioxide (ppm - calculated by the TVOC sensor)",
		[]string{
			"device_uuid",
		},
		nil,
	)

	co2_estimate_baseline = prometheus.NewDesc(
		prometheus.BuildFQName("awair", "", "co2_est_baseline"),
		"A unitless value that represents the baseline from which the TVOC sensor partially derives its estimated (e)CO₂output.",
		[]string{
			"device_uuid",
		},
		nil,
	)

	voc = prometheus.NewDesc(
		prometheus.BuildFQName("awair", "", "voc"),
		"Total Volatile Organic Compounds (ppb)",
		[]string{
			"device_uuid",
		},
		nil,
	)

	voc_baseline = prometheus.NewDesc(
		prometheus.BuildFQName("awair", "", "voc_baseline"),
		"A unitless value that represents the baseline from which the TVOC sensor partially derives its TVOC output.",
		[]string{
			"device_uuid",
		},
		nil,
	)

	voc_h2_raw = prometheus.NewDesc(
		prometheus.BuildFQName("awair", "", "voc_h2_raw"),
		"A unitless value that represents the Hydrogen gas signal from which the TVOC sensor partially derives its TVOC output.",
		[]string{
			"device_uuid",
		},
		nil,
	)

	voc_ethanol_raw = prometheus.NewDesc(
		prometheus.BuildFQName("awair", "", "voc_ethanol_raw"),
		"A unitless value that represents the Ethanol gas signal from which the TVOC sensor partially derives its TVOC output.",
		[]string{
			"device_uuid",
		},
		nil,
	)

	pm25 = prometheus.NewDesc(
		prometheus.BuildFQName("awair", "", "pm25"),
		"Particulate matter less than 2.5 microns in diameter (µg/m³)",
		[]string{
			"device_uuid",
		},
		nil,
	)

	pm10 = prometheus.NewDesc(
		prometheus.BuildFQName("awair", "", "pm10"),
		"Estimated particulate matter less than 10 microns in diameter (µg/m³ - calculated by the PM2.5 sensor)",
		[]string{
			"device_uuid",
		},
		nil,
	)
	info = prometheus.NewDesc(
		prometheus.BuildFQName("awair", "", "device_info"),
		"Info about the awair device",
		[]string{
			"device_uuid",
			"firmware_version",
			"voc_feature_set",
		},
		nil,
	)
)

type AwairValues struct {
	Score          float64 `json:"score"`
	DewPoint       float64 `json:"dew_point"`
	Temp           float64 `json:"temp"`
	Humidity       float64 `json:"humid"`
	AbsHumidity    float64 `json:"abs_humid"`
	CO2            float64 `json:"co2"`
	CO2Est         float64 `json:"co2_est"`
	CO2EstBaseline float64 `json:"co2_est_baseline"`
	Voc            float64 `json:"voc"`
	VocBaseline    float64 `json:"voc_baseline"`
	VocH2Raw       float64 `json:"voc_h2_raw"`
	VocEthanolRaw  float64 `json:"voc_ethanol_raw"`
	PM25           float64 `json:"pm25"`
	PM10Est        float64 `json:"pm10_est"`
}

type LEDSettings struct {
	Mode       string
	Brightness int
}

type ConfigResponse struct {
	DeviceUUID      string      `json:"device_uuid"`
	WifiMAC         string      `json:"wifi_mac"`
	SSID            string      `json:"ssid"`
	IP              string      `json:"ip"`
	Netmask         string      `json:"netmask"`
	Gateway         string      `json:"gateway"`
	FirmwareVersion string      `json:"fw_version"`
	Timezone        string      `json:"timezone"`
	Display         string      `json:"display"`
	LED             LEDSettings `json:"led"`
	VocFeatureSet   int         `json:"voc_feature_set"`
}

type AwairExporter struct {
	hostname string
}

func NewAwairExporter(hostname string) (*AwairExporter, error) {
	ex := &AwairExporter{
		hostname: hostname,
	}
	config, err := ex.GetConfig()
	if err != nil {
		return nil, err
	}
	log.Info().
		Interface("config", config).
		Msg("Successfully connected to Awair device.")

	return ex, nil
}

func (e *AwairExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- score
	ch <- dew_point
	ch <- temp
	ch <- humidity
	ch <- abs_humidity
	ch <- co2
	ch <- co2_estimated
	ch <- co2_estimate_baseline
	ch <- voc
	ch <- voc_baseline
	ch <- voc_h2_raw
	ch <- voc_ethanol_raw
	ch <- pm25
	ch <- pm10
	ch <- info
}

func (e *AwairExporter) GetMetrics() (*AwairValues, error) {
	uri := fmt.Sprintf("http://%s/air-data/latest", e.hostname)
	log.Debug().
		Str("uri", uri).
		Msg("Attempting to retrieve metrics from Awair device.")

	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	values := AwairValues{}
	err = json.Unmarshal(body, &values)
	if err != nil {
		return nil, err
	}
	return &values, nil
}

func (e *AwairExporter) GetConfig() (*ConfigResponse, error) {
	uri := fmt.Sprintf("http://%s/settings/config/data", e.hostname)
	log.Debug().
		Str("uri", uri).
		Msg("Attempting to retrieve config from Awair device.")

	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	config := ConfigResponse{}
	err = json.Unmarshal(body, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (e *AwairExporter) Collect(ch chan<- prometheus.Metric) {
	values := &AwairValues{}
	config := &ConfigResponse{}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		var err error
		values, err = e.GetMetrics()
		if err != nil {
			log.Error().Err(err).
				Msg("Error retrieving Metrics from device")
		}
		wg.Done()
		log.Debug().
			Interface("metrics", values).
			Msg("Metrics successfully retrieved")
	}()
	go func() {
		var err error
		config, err = e.GetConfig()
		if err != nil {
			log.Error().Err(err).
				Msg("Error retrieving Metrics from device")
		}
		log.Debug().
			Interface("config", config).
			Msg("Config successfully retrieved")
		wg.Done()
	}()
	wg.Wait()

	ch <- prometheus.MustNewConstMetric(
		score, prometheus.GaugeValue, values.Score, config.DeviceUUID,
	)
	ch <- prometheus.MustNewConstMetric(
		dew_point, prometheus.GaugeValue, values.DewPoint, config.DeviceUUID,
	)
	ch <- prometheus.MustNewConstMetric(
		temp, prometheus.GaugeValue, values.Temp, config.DeviceUUID,
	)
	ch <- prometheus.MustNewConstMetric(
		humidity, prometheus.GaugeValue, values.Humidity, config.DeviceUUID,
	)
	ch <- prometheus.MustNewConstMetric(
		abs_humidity, prometheus.GaugeValue, values.AbsHumidity, config.DeviceUUID,
	)
	ch <- prometheus.MustNewConstMetric(
		co2, prometheus.GaugeValue, values.CO2, config.DeviceUUID,
	)
	ch <- prometheus.MustNewConstMetric(
		co2_estimated, prometheus.GaugeValue, values.CO2Est, config.DeviceUUID,
	)
	ch <- prometheus.MustNewConstMetric(
		co2_estimate_baseline, prometheus.GaugeValue, values.CO2EstBaseline, config.DeviceUUID,
	)
	ch <- prometheus.MustNewConstMetric(
		voc, prometheus.GaugeValue, values.Voc, config.DeviceUUID,
	)
	ch <- prometheus.MustNewConstMetric(
		voc_baseline, prometheus.GaugeValue, values.VocBaseline, config.DeviceUUID,
	)
	ch <- prometheus.MustNewConstMetric(
		voc_h2_raw, prometheus.GaugeValue, values.VocH2Raw, config.DeviceUUID,
	)
	ch <- prometheus.MustNewConstMetric(
		voc_ethanol_raw, prometheus.GaugeValue, values.VocEthanolRaw, config.DeviceUUID,
	)
	ch <- prometheus.MustNewConstMetric(
		pm25, prometheus.GaugeValue, values.PM25, config.DeviceUUID,
	)
	ch <- prometheus.MustNewConstMetric(
		pm10, prometheus.GaugeValue, values.PM10Est, config.DeviceUUID,
	)
	ch <- prometheus.MustNewConstMetric(
		info, prometheus.GaugeValue, 1,
		config.DeviceUUID,
		config.FirmwareVersion,
		strconv.Itoa(config.VocFeatureSet),
	)
}
