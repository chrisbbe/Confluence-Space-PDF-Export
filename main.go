package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var loginPayload = `<x:Envelope xmlns:x="http://schemas.xmlsoap.org/soap/envelope/" xmlns:rpc="http://rpc.confluence.atlassian.com">
    <x:Header/>
    <x:Body>
        <rpc:login>
            <rpc:in0>%s</rpc:in0>
            <rpc:in1>%s</rpc:in1>
        </rpc:login>
    </x:Body>
</x:Envelope>`

var downloadPayload = `<x:Envelope xmlns:x="http://schemas.xmlsoap.org/soap/envelope/" xmlns:rpc1="http://rpc.flyingpdf.extra.confluence.atlassian.com">
    <x:Header/>
    <x:Body>
        <rpc1:exportSpace>
            <rpc1:in0>%s</rpc1:in0>
            <rpc1:in1>%s</rpc1:in1>
        </rpc1:exportSpace>
    </x:Body>
</x:Envelope>`

func main() {
	confluenceServerBaseURL := flag.String("server", "", "URL to Confluence server")
	confluenceUsername := flag.String("username", "", "Confluence username for user to perform export")
	confluencePassword := flag.String("password", "", "Confluence password for user to perform export")
	confluenceSpace := flag.String("spaceKey", "", "Confluence space key to export")
	exportFileLocation := flag.String("exportDirectory", "", "Fully qualified file location to store export")

	flag.Parse()
	if *confluenceServerBaseURL == "" || *confluenceUsername == "" || *confluencePassword == "" || *confluenceSpace == "" || *exportFileLocation == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Check that local folder exist.
	if src, err := os.Stat(*exportFileLocation); src != nil && !src.IsDir() || os.IsNotExist(err) {
		log.Fatalf("Path %s either does not exist or is a file", *exportFileLocation)
	}

	log.Printf("Start Confluence Space PDF export")
	token, err := getConfluenceLoginToken(*confluenceServerBaseURL, *confluenceUsername, *confluencePassword)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Login successfull")
	log.Printf("Export started...")

	spacePdfURL, err := GetConfluenceSpaceExportURL(*confluenceServerBaseURL, token, *confluenceSpace)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Export completed.")
	log.Printf("Downloading PDF...")

	if err := DownloadFile(spacePdfURL, *confluenceUsername, *confluencePassword, *confluenceSpace, *exportFileLocation); err != nil {
		log.Fatal(err)
	}
	log.Printf("PDF downloaded successfully.")
}

type LoginResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Text    string   `xml:",chardata"`
	Soapenv string   `xml:"soapenv,attr"`
	Xsd     string   `xml:"xsd,attr"`
	Xsi     string   `xml:"xsi,attr"`
	Body    struct {
		Text          string `xml:",chardata"`
		LoginResponse struct {
			Text          string `xml:",chardata"`
			EncodingStyle string `xml:"encodingStyle,attr"`
			Ns1           string `xml:"ns1,attr"`
			LoginReturn   struct {
				Text string `xml:",chardata"`
				Type string `xml:"type,attr"`
			} `xml:"loginReturn"`
		} `xml:"loginResponse"`
	} `xml:"Body"`
}

func getConfluenceLoginToken(confluenceServerBaseURL, username, password string) (string, error) {
	request, err := http.NewRequest(http.MethodPost, confluenceServerBaseURL+"/plugins/servlet/soap-axis1/pdfexport", strings.NewReader(fmt.Sprintf(loginPayload, username, password)))
	if err != nil {
		return "", err
	}

	request.Header.Add("Content-Type", "text/xml; charset=utf-8")
	request.Header.Add("SOAPAction", "")

	response, err := (&http.Client{}).Do(request)
	if err != nil {
		return "", err
	}
	defer Close(response.Body)

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status is %d, should be %d", response.StatusCode, http.StatusOK)
	}

	var loginResponse LoginResponse
	if err := xml.NewDecoder(response.Body).Decode(&loginResponse); err != nil {
		return "", err
	}

	return loginResponse.Body.LoginResponse.LoginReturn.Text, nil
}

type ExportSpaceResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Text    string   `xml:",chardata"`
	Soapenv string   `xml:"soapenv,attr"`
	Xsd     string   `xml:"xsd,attr"`
	Xsi     string   `xml:"xsi,attr"`
	Body    struct {
		Text                string `xml:",chardata"`
		ExportSpaceResponse struct {
			Text              string `xml:",chardata"`
			EncodingStyle     string `xml:"encodingStyle,attr"`
			Ns1               string `xml:"ns1,attr"`
			ExportSpaceReturn struct {
				Text string `xml:",chardata"`
				Type string `xml:"type,attr"`
			} `xml:"exportSpaceReturn"`
		} `xml:"exportSpaceResponse"`
	} `xml:"Body"`
}

func GetConfluenceSpaceExportURL(confluenceServerBaseURL, loginToken, spaceKey string) (string, error) {
	request, err := http.NewRequest(http.MethodPost, confluenceServerBaseURL+"/plugins/servlet/soap-axis1/pdfexport", strings.NewReader(fmt.Sprintf(downloadPayload, loginToken, spaceKey)))
	if err != nil {
		return "", err
	}
	request.Header.Add("Content-Type", "text/xml; charset=utf-8")
	request.Header.Add("SOAPAction", "")

	response, err := (&http.Client{}).Do(request)
	if err != nil {
		return "", err
	}
	defer Close(response.Body)

	var exportResponse ExportSpaceResponse
	if err := xml.NewDecoder(response.Body).Decode(&exportResponse); err != nil {
		return "", err
	}

	return exportResponse.Body.ExportSpaceResponse.ExportSpaceReturn.Text, nil
}

func DownloadFile(url, username, password, spaceKey, exportLocation string) error {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	request.SetBasicAuth(username, password)

	response, err := (&http.Client{}).Do(request)
	if err != nil {
		return err
	}
	defer Close(response.Body)

	out, err := os.Create(filepath.FromSlash(fmt.Sprintf("%s/%s-%s.pdf", exportLocation, spaceKey, time.Now().Format("2006-01-02-150405"))))
	if err != nil {
		return err
	}
	defer Close(out)

	_, err = io.Copy(out, response.Body)
	return err
}

func Close(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Fatal(err)
	}
}
