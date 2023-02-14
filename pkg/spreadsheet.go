package pkg

import (
	"io/ioutil"

	"github.com/apex/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gopkg.in/Iwark/spreadsheet.v2"
)

const SPREADSHEET_ID = "1o82xzrBiKeCmoiUB5iXxBy2gyP1g0Spu-FT8Ig-O50U"

type SpreadsheetClub struct {
	Service     *spreadsheet.Service
	Spreadsheet *spreadsheet.Spreadsheet
}

func GetSheet() *SpreadsheetClub {
	data, err := ioutil.ReadFile("./config/client_secret.json")
	if err != nil {
		log.Error(err.Error())
	}

	conf, err := google.JWTConfigFromJSON(data, spreadsheet.Scope)
	if err != nil {
		log.Error(err.Error())
	}

	client := conf.Client(oauth2.NoContext)

	service := spreadsheet.NewServiceWithClient(client)

	spreadsheet, err := service.FetchSpreadsheet(SPREADSHEET_ID)
	if err != nil {
		log.Error(err.Error())
	}

	log.Infof("get spreadsheet success. ID=%s", spreadsheet.ID)

	return &SpreadsheetClub{
		Service:     service,
		Spreadsheet: &spreadsheet,
	}
}

func (spreadsheetClub *SpreadsheetClub) Reload() {
	spreadsheet, err := spreadsheetClub.Service.FetchSpreadsheet(SPREADSHEET_ID)
	if err != nil {
		log.Error(err.Error())
	}

	spreadsheetClub.Spreadsheet = &spreadsheet

	log.Infof("get spreadsheet success. ID=%s", spreadsheet.ID)
}
