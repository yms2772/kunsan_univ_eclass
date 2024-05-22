package api

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type tkisAllScoreResponse struct {
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
				Text string `xml:",chardata"`
				ID   string `xml:"id,attr"`
				Type string `xml:"type,attr"`
				Size string `xml:"size,attr"`
			} `xml:"Column"`
		} `xml:"ColumnInfo"`
		Rows struct {
			Text string `xml:",chardata"`
			Row  []struct {
				Text string `xml:",chardata"`
				Col  []struct {
					Text string `xml:",chardata"`
					ID   string `xml:"id,attr"`
				} `xml:"Col"`
			} `xml:"Row"`
		} `xml:"Rows"`
	} `xml:"Dataset"`
}

type SubjectScoreData struct {
	Name  string
	Score float64
	Grade string
}

type ScoreData struct {
	Data [4][2][]SubjectScoreData
}

func (u *user) GetScore() (ScoreData, error) {
	var data ScoreData

	reqStr := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?><Root xmlns="http://www.nexacroplatform.com/platform/dataset"><Parameters><Parameter id="WMONID"></Parameter><Parameter id="fsp_action">xDefaultAction</Parameter><Parameter id="fsp_cmd">execute</Parameter><Parameter id="GV_USER_ID">%s</Parameter><Parameter id="GV_IP_ADDRESS"></Parameter><Parameter id="fsp_logId">all</Parameter></Parameters><Dataset id="ds_cond"><ColumnInfo><Column id="HAKBUN" type="STRING" size="256"  /></ColumnInfo><Rows><Row><Col id="HAKBUN">%s</Col></Row></Rows></Dataset><Dataset id="fsp_ds_cmd"><ColumnInfo><Column id="TX_NAME" type="string" size="100"  /><Column id="TYPE" type="string" size="10"  /><Column id="SQL_ID" type="string" size="200"  /><Column id="KEY_SQL_ID" type="string" size="200"  /><Column id="KEY_INCREMENT" type="INT" size="10"  /><Column id="CALLBACK_SQL_ID" type="STRING" size="200"  /><Column id="INSERT_SQL_ID" type="STRING" size="200"  /><Column id="UPDATE_SQL_ID" type="STRING" size="200"  /><Column id="DELETE_SQL_ID" type="STRING" size="200"  /><Column id="SAVE_FLAG_COLUMN" type="STRING" size="200"  /><Column id="USE_INPUT" type="STRING" size="1"  /><Column id="USE_ORDER" type="STRING" size="1"  /><Column id="KEY_ZERO_LEN" type="INT" size="10"  /><Column id="BIZ_NAME" type="STRING" size="100"  /><Column id="PAGE_NO" type="INT" size="10"  /><Column id="PAGE_SIZE" type="INT" size="10"  /><Column id="READ_ALL" type="STRING" size="1"  /><Column id="EXEC_TYPE" type="STRING" size="2"  /><Column id="EXEC" type="STRING" size="1"  /><Column id="FAIL" type="STRING" size="1"  /><Column id="FAIL_MSG" type="STRING" size="200"  /><Column id="EXEC_CNT" type="INT" size="1"  /><Column id="MSG" type="STRING" size="200"  /></ColumnInfo><Rows><Row><Col id="TX_NAME" /><Col id="TYPE">N</Col><Col id="SQL_ID">unaf/scor/sd:SCORSD010_S01</Col><Col id="KEY_SQL_ID" /><Col id="KEY_INCREMENT">0</Col><Col id="CALLBACK_SQL_ID" /><Col id="INSERT_SQL_ID" /><Col id="UPDATE_SQL_ID" /><Col id="DELETE_SQL_ID" /><Col id="SAVE_FLAG_COLUMN" /><Col id="USE_INPUT" /><Col id="USE_ORDER" /><Col id="KEY_ZERO_LEN">0</Col><Col id="BIZ_NAME" /><Col id="PAGE_NO" /><Col id="PAGE_SIZE" /><Col id="READ_ALL" /><Col id="EXEC_TYPE">B</Col><Col id="EXEC" /><Col id="FAIL" /><Col id="FAIL_MSG" /><Col id="EXEC_CNT">0</Col><Col id="MSG" /></Row></Rows></Dataset></Root>`,
		u.id,
		u.id,
	)

	req, err := http.NewRequest(http.MethodPost, "https://tkis.kunsan.ac.kr/XMain", strings.NewReader(reqStr))
	if err != nil {
		return data, err
	}

	req.Header.Set("Content-Type", "application/xml")
	req.Header.Set("Referer", "https://tkis.kunsan.ac.kr/index.do")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	resp, err := u.getClient().Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil || resp == nil {
		return data, err
	}

	var respData tkisAllScoreResponse
	if err := xml.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return data, err
	}

	if len(respData.Dataset) == 0 {
		return data, fmt.Errorf("데이터가 없습니다")
	}

	for _, row := range respData.Dataset[0].Rows.Row {
		year := 0
		semester := 0
		name := ""
		grade := ""
		score := float64(0)

		for _, col := range row.Col {
			switch col.ID {
			case "GMNAME":
				name = col.Text
			case "HAKYOUN":
				n, err := strconv.Atoi(col.Text)
				if err != nil {
					year = -1
					continue
				}
				year = n
			case "HAKGI":
				n, err := strconv.Atoi(col.Text)
				if err != nil {
					semester = -1
					continue
				}
				semester = n
			case "WHANJUMSU":
				grade = col.Text
			case "CHEJUMSU":
				score, _ = strconv.ParseFloat(col.Text, 10)
			}
		}

		if year == -1 || semester == -1 {
			continue
		}

		data.Data[year-1][semester-1] = append(data.Data[year-1][semester-1], SubjectScoreData{
			Name:  name,
			Score: score,
			Grade: grade,
		})
	}
	return data, nil
}
