package api

import "encoding/xml"

type tkisResponse struct {
	XMLName    xml.Name `xml:"Root"`
	Text       string   `xml:",chardata"`
	Xmlns      string   `xml:"xmlns,attr"`
	Ver        string   `xml:"ver,attr"`
	Parameters struct {
		Text      string `xml:",chardata"`
		Parameter []struct {
			Text string `xml:",chardata"`
			ID   string `xml:"id,attr"`
			Type string `xml:"type,attr"`
		} `xml:"Parameter"`
	} `xml:"Parameters"`
	Dataset []struct {
		Text       string `xml:",chardata"`
		ID         string `xml:"id,attr"`
		ColumnInfo struct {
			Text   string `xml:",chardata"`
			Column []struct {
				Text    string `xml:",chardata"`
				ID      string `xml:"id,attr"`
				Type    string `xml:"type,attr"`
				Size    string `xml:"size,attr"`
				Encrypt string `xml:"encrypt,attr"`
			} `xml:"Column"`
		} `xml:"ColumnInfo"`
		Rows struct {
			Text string `xml:",chardata"`
			Row  struct {
				Text string `xml:",chardata"`
				Col  []struct {
					Text string `xml:",chardata"`
					ID   string `xml:"id,attr"`
				} `xml:"Col"`
			} `xml:"Row"`
		} `xml:"Rows"`
	} `xml:"Dataset"`
}
