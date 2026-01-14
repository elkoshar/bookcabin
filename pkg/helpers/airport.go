package helpers

import "strings"

type AirportInfo struct {
	Code     string
	City     string
	Name     string
	Timezone string
}

// AirportMap contains only airports that exist in the current mock data
var AirportMap = map[string]AirportInfo{
	"CGK": {Code: "CGK", City: "Jakarta", Name: "Soekarno-Hatta International Airport", Timezone: "WIB"},
	"DPS": {Code: "DPS", City: "Denpasar", Name: "I Gusti Ngurah Rai International Airport", Timezone: "WITA"},
	"SUB": {Code: "SUB", City: "Surabaya", Name: "Juanda International Airport", Timezone: "WIB"},
}

func GetAirportDetail(code string) AirportInfo {
	code = strings.ToUpper(code)
	if info, ok := AirportMap[code]; ok {
		return info
	}
	return AirportInfo{
		Code:     code,
		City:     code,
		Name:     "Unknown Airport",
		Timezone: "WIB",
	}
}

func GetCityName(code string) string {
	return GetAirportDetail(code).City
}
