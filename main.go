package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

type Envelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		XMLName      xml.Name `xml:"Body"`
		ExchangeRate struct {
			XMLName xml.Name `xml:"getLatestInterestAndExchangeRatesResponse"`
			Groups  struct {
				//XMLName xml.Name `xml:"return>groups"`
				GroupId   string `xml:"groupid"`
				GroupName string `xml:"groupname"`
				Series    []struct {
					SeriesId   string  `xml:"seriesid"`
					SeriesName string  `xml:"seriesname"`
					Unit       float32 `xml:"unit"`
					Value      float32 `xml:"resultrows>value"`
				} `xml:"series"`
			} `xml:"return>groups"`
		}
	}
}

const getEnvelope = `<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:xsd="http://swea.riksbank.se/xsd">
	<soap:Body>
	<xsd:getLatestInterestAndExchangeRates>
        	<languageid>en</languageid>
        	<seriesid>{{.CurrencyCode}}</seriesid>
      </xsd:getLatestInterestAndExchangeRates>   
      </soap:Body>
</soap:Envelope>`

func main() {
	http.HandleFunc("/", MainHandler)
	log.Println(http.ListenAndServe(":8080", nil))
}

func MainHandler(w http.ResponseWriter, r *http.Request) {
	rate, err := getCurrencyRate("SEKEURPMI")
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Fprintf(w, fmt.Sprintf("%f", rate))
}

func getCurrencyRate(currencyCode string) (float32, error) {
	tmpl, err := template.New("getEnvelope").Parse(getEnvelope)
	if err != nil {
		return 0, err
	}
	var docs bytes.Buffer
	type CurrencyData struct {
		CurrencyCode string
	}
	currency := CurrencyData{CurrencyCode: currencyCode}
	if err := tmpl.Execute(&docs, currency); err != nil {
		return 0, err
	}
	url := "https://swea.riksbank.se:443/sweaWS/services/SweaWebServiceHttpSoap12Endpoint"
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, &docs)
	if err != nil {
		return 0, err
	}
	req.Header.Add("Content-Type", "application/xml")
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	//fmt.Println(string(contents))

	test := Envelope{}
	if err := xml.Unmarshal(contents, &test); err != nil {
		return 0, err
	}
	return test.Body.ExchangeRate.Groups.Series[0].Value, nil
}
