package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

const url = "https://api.coinlore.net/api/ticker/?id="

// toks rules.json failo kelias ihardkodinimas privercia visada programa
// paleisti is root direktorijos kaip `go run cmd/main.go`. Jei bandysi paleisti
// kaip `cd cmd/ && go run main.go`, programa neras rules.json failo.
const rulesFile = "./assets/rules.json"

type RuleStruct struct {
	RulesArray []Rule `json:"rules"`
}

type Rule struct {
	// reiketu panaudoti go fmt ir panasius irankius rasant koda. Editorius
	// butu praneses kad go nelabai megsta "_" kintamuju pavadinimuose. Taip
	// pat visokie akronimai kaip id, http, usd ir pan turetu but didziosiomis
	// raidemis: ID, HTTP, USD. Sia informacija galiam rast aprasytus prie
	// go kodo standartu.
	Crypto_id string `json:"crypto_id"`
	Price     string `json:"price"`
	Rule      string `json:"rule"`
}

type APIResult struct {
	Id               string `json:"id"`
	Symbol           string `json:"symbol"`
	Name             string `json:"name"`
	NameId           string `json:"nameid"`
	Rank             int    `json:"rank"`
	Price_usd        string `json:"price_usd"`
	Percent_change24 string `json:"percent_change_24h"`
	Percent_change1  string `json:"percent_change_1h"`
	Percent_change7d string `json:"percent_change_7d"`
	Market_cap       string `json:"market_cap_usd"`
	Volume24         string `json:"volume24"`
	Volume24_native  string `json:"volume24_native"`
	Csupply          string `json:"csupply"`
	Price_btc        string `json:"price_btc"`
	Tsupply          string `json:"tsupply"`
	Msupply          string `json:"msupply"`
}

type shortAPI struct {
	Id    string
	Name  string
	Price string
}

// siuo atveju, butu geriau jei f funkcijos parasas butu  aprasytas kaip atskiras type.
// context.Context, jei imanoma visada turetu vadintis ctx. Cia panasi taisykle, kaip for
// cikluose pirma bandoma naudoti i,j raides ir pan., arba kaip error tipa laikantis kintamasis
// turetu vadintis err.
func doEvery(cText context.Context, d time.Duration, f func(string, RuleStruct, time.Time) RuleStruct) error {
	// tickerius visada reikia sustabdyti. Ta galima padaryti su defer ticker.Stop()
	// arba jei netinka su defer, suhandlinti pries return.
	ticker := time.Tick(1)
	var structure RuleStruct
	for {
		select {
		case <-cText.Done():
			return cText.Err()
		case x := <-ticker:
			// cia kazkas labai negerai. Reiketu tikeri paleist iskarto su d
			// ir vidury "tickinimo" pernaujo tickerio nustatyti nereiketu.
			ticker = time.Tick(d)
			// cia kazkas negerai. perdaug yra pernaudojamas structure kintamasis.
			// galima panaudoti pointerius, tam kad nereiketu taip pernaudot structure
			// kintamojo. Tuomet funkijos galetu structure vidurius pakoreguot iskart.
			structure := f(rulesFile, structure, x)
			for i := 0; i < len(structure.RulesArray); i++ {
				fmt.Printf("Crypto_id:%v\n", structure.RulesArray[i].Crypto_id)
			}
			structure = checkRules(structure)
			reWritefile(structure)
		}

	}
}

func ReadFile(fileName string, rules RuleStruct, t time.Time) RuleStruct {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
	}

	defer file.Close()
	fmt.Printf("File successfully opened at %v\n", t)

	// cia kazkas negerai. Keistai atrodo kad atidarai faila su os.Open, bet
	// nuskaitymui naudoji ioutil biblioteka. Jei jau naudoji ioutil, tuomet 
	// visa koda auksciau galetum ismesti jei panaudotum ioutil.ReadFile(path)
	byteValue, _ := ioutil.ReadAll(file)
	
	json.Unmarshal(byteValue, &rules)
	return rules
}

func checkRules(structure RuleStruct) RuleStruct {

	// sitas kintamasis nereikalingas jei naudotum
	// for i, v := range structure.RulesArray { .. }
	array := structure.RulesArray
	for i := 0; i < len(array); i++ {
		isUsed := false
		result := GetAPI(array[i].Crypto_id)
		isUsed = findAPIresult(result, array[i], isUsed)
		if isUsed {
			// sitas neveiks taip kaip tu tikiesi. e.g.
			// https://play.golang.org/p/Y7jyzTS4tJZ
			array[0], array[i] = array[i], array[0]
			structure.RulesArray = array[1:]

		}
	}
	return structure
}

func findAPIresult(API shortAPI, rule Rule, IsUsed bool) bool {
	floatRulePrice, _ := strconv.ParseFloat(rule.Price, 64)
	floatApiPrice, _ := strconv.ParseFloat(API.Price, 64)
	switch rule.Rule {
	case "lt":
		if floatApiPrice < floatRulePrice && !IsUsed {
			fmt.Printf("Cryptocurrency id:"+API.Id+" "+API.Name+" is lower than %v\n", rule.Price)
			IsUsed = true
		}
	case "gt":
		if floatApiPrice > floatRulePrice && !IsUsed {
			fmt.Printf("Cryptocurrency id:"+API.Id+" "+API.Name+" is grater than %v\n", rule.Price)
			IsUsed = true
		}

	}

	// nereiketu pernaudot to paties kintamojo cia, kuri gavai
	// per funkcijos argumentus.
	return IsUsed

}

func GetAPI(Id string) shortAPI {

	URL := fmt.Sprintf(url + Id)
	response, err := http.Get(URL)
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	// truksta defer response.Body.Close()
	responseData, _ := ioutil.ReadAll(response.Body)

	//responseData[:len(responseData)] =125
	//trimedResponseData :=responseData[0:]
	//trimedResponseData2:=trimedResponseData[:len(trimedResponseData)-1]

	//var responseObject APIStruct
	var responseObject []APIResult
	error := json.Unmarshal(responseData, &responseObject)

	if error != nil {
		fmt.Printf("There was an error decoding the json. err = %s", error)
	}
	//	oneAPI :=responseObject.ApiArray[0]
	//	result := createShortApi(oneAPI.Id,oneAPI.Name,oneAPI.Price_usd)
	
	// jei ivyko klaida, responseObject bus tuscias array, todel tokio elemento
	// kaip [0] nepavyks pasiekti. Potenciali vieta gauti panic.
	result := createShortApi(responseObject[0].Id, responseObject[0].Name, responseObject[0].Price_usd)
	return result
}

func reWritefile(structure RuleStruct) {
	file, _ := json.MarshalIndent(structure, "", " ")

	// matai, failo irasymui panaudoji WriteFile, be jokiu os.Open(..),
	// kodel nuskaitymui nenaudoji irgi tiesiog ReadFile? :D
	_ = ioutil.WriteFile("./assets/rules.json", file, 0644)

}

func createShortApi(id string, name string, price string) shortAPI {
	API := shortAPI{Id: id, Name: name, Price: price}
	return API
}

func main() {
	fmt.Println()
	// kodel cia naudoji WithTimeout? Tavo programa nedirbs ilgiau kaip 5 min.
	// ar tikrai ta nori pasiekti cia?
	cText, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// tokios funkcijos kurios viduje turi tickerius, turetu but paleidziamos
	// atskiroje rutinoje kaip go doEvery(..).
	doEvery(cText, 30*time.Second, ReadFile)

}
