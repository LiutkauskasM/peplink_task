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
	"sync"
)

const url = "https://api.coinlore.net/api/ticker/?id="

// toks rules.json failo kelias ihardkodinimas privercia visada programa
// paleisti is root direktorijos kaip `go run cmd/main.go`. Jei bandysi paleisti
// kaip `cd cmd/ && go run main.go`, programa neras rules.json failo.
const rulesFile = "./assets/rules.json"
var wg sync.WaitGroup



//naudojamas nskaityti visoms taiyklems is rules.json failo
type RuleStruct struct {
	RulesArray []Rule `json:"rules"`
}


// struktura rules.json objektams saugoti, kurioje yra visos taisykles savybes
type Rule struct {
	// reiketu panaudoti go fmt ir panasius irankius rasant koda. Editorius
	// butu praneses kad go nelabai megsta "_" kintamuju pavadinimuose. Taip
	// pat visokie akronimai kaip id, http, usd ir pan turetu but didziosiomis
	// raidemis: ID, HTTP, USD. Sia informacija galiam rast aprasytus prie
	// go kodo standartu.
	Crypto_ID string `json:"crypto_id"`
	Price     string `json:"price"`
	Rule      string `json:"rule"`
}
// struktura, skirta saugoti Is API gautiems rezultatams
type APIResult struct {
	Id               string `json:"id"`
	Symbol           string `json:"symbol"`
	Name             string `json:"name"`
	NameId           string `json:"nameid"`
	Rank             int    `json:"rank"`
	Price_USD        string `json:"price_usd"`
	Percent_change24 string `json:"percent_change_24h"`
	Percent_change1  string `json:"percent_change_1h"`
	Percent_change7d string `json:"percent_change_7d"`
	Market_cap       string `json:"market_cap_usd"`
	Volume24         string `json:"volume24"`
	Volume24_native  string `json:"volume24_native"`
	Csupply          string `json:"csupply"`
	Price_BTC        string `json:"price_btc"`
	Tsupply          string `json:"tsupply"`
	Msupply          string `json:"msupply"`
}


// siuo atveju, butu geriau jei f funkcijos parasas butu  aprasytas kaip atskiras type.
// context.Context, jei imanoma visada turetu vadintis ctx. Cia panasi taisykle, kaip for
// cikluose pirma bandoma naudoti i,j raides ir pan., arba kaip error tipa laikantis kintamasis
// turetu vadintis err.
func doEvery(ctx context.Context, d time.Duration, f func(string, *RuleStruct, time.Time)) error {
	// tickerius visada reikia sustabdyti. Ta galima padaryti su defer ticker.Stop()
	// arba jei netinka su defer, suhandlinti pries return.
	ticker := time.NewTicker(d)
	defer ticker.Stop()
	//defer time.Stop(&ticker)
	var structure RuleStruct
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case x := <-ticker.C:
			// cia kazkas labai negerai. Reiketu tikeri paleist iskarto su d
			// ir vidury "tickinimo" pernaujo tickerio nustatyti nereiketu.
		
			// cia kazkas negerai. perdaug yra pernaudojamas structure kintamasis.
			// galima panaudoti pointerius, tam kad nereiketu taip pernaudot structure
			// kintamojo. Tuomet funkijos galetu structure vidurius pakoreguot iskart.
			f(rulesFile, &structure, x)
			for i, rule := range structure.RulesArray {
				fmt.Printf("#%d Crypto_id:%v\n",i+1, rule.Crypto_ID)
			}
			checkRules(&structure)
			reWritefile(structure)
		}

	}
}

func readFile(fileName string, rules *RuleStruct, t time.Time) {
	//file, err := os.Open(fileName)
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("File successfully opened at %v\n", t)

	// cia kazkas negerai. Keistai atrodo kad atidarai faila su os.Open, bet
	// nuskaitymui naudoji ioutil biblioteka. Jei jau naudoji ioutil, tuomet 
	// visa koda auksciau galetum ismesti jei panaudotum ioutil.ReadFile(path)

	//byteValue, _ := ioutil.ReadAll(file)
	
	err =	json.Unmarshal(file, &rules)
	if err != nil {
		fmt.Printf("There was an error decoding the json. err = %s", err)
		os.Exit(1)
	}
	
}

func checkRules(structure *RuleStruct){

	// sitas kintamasis nereikalingas jei naudotum
	// for i, v := range structure.RulesArray { .. }
	for i, rule := range structure.RulesArray {
		result := getAPI(rule.Crypto_ID)
		isUsed := findAPIresult(result, rule)
		if isUsed {
			// sitas neveiks taip kaip tu tikiesi. e.g.
			// https://play.golang.org/p/Y7jyzTS4tJZ
			//M: bet kai masyvas vis pildomas, turėtų veikti, nes failas vis nuskaitomas
			// ir gali būti skirtingos taisyklės ( labiau gal toks variantas būtų : https://play.golang.org/p/PElaOgE0YQu)
			structure.RulesArray[0], structure.RulesArray[i] = structure.RulesArray[i], structure.RulesArray[0]
			structure.RulesArray = structure.RulesArray[1:]

		}
	}
}

func findAPIresult(API APIResult, rule Rule) bool {
	floatRulePrice, _ := strconv.ParseFloat(rule.Price, 64)
	floatApiPrice, _ := strconv.ParseFloat(API.Price_USD, 64)
	IsUsed:= false
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

func getAPI(Id string) APIResult {

	URL := fmt.Sprintf(url + Id)
	response, err := http.Get(URL)
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	// truksta defer response.Body.Close()
	defer response.Body.Close()
	responseData, _ := ioutil.ReadAll(response.Body)

	//responseData[:len(responseData)] =125
	//trimedResponseData :=responseData[0:]
	//trimedResponseData2:=trimedResponseData[:len(trimedResponseData)-1]

	//var responseObject APIStruct
	var responseObject []APIResult
	err = json.Unmarshal(responseData, &responseObject)

	if err != nil {
		fmt.Printf("There was an error decoding the json. err = %s", err)
		os.Exit(1)
	}
	//	oneAPI :=responseObject.ApiArray[0]
	//	result := createShortApi(oneAPI.Id,oneAPI.Name,oneAPI.Price_usd)
	
	// jei ivyko klaida, responseObject bus tuscias array, todel tokio elemento
	// kaip [0] nepavyks pasiekti. Potenciali vieta gauti panic.
	return responseObject[0]
}

func reWritefile(structure RuleStruct) {
	file, _ := json.MarshalIndent(structure, "", " ")

	// matai, failo irasymui panaudoji WriteFile, be jokiu os.Open(..),
	// kodel nuskaitymui nenaudoji irgi tiesiog ReadFile? :D
	_ = ioutil.WriteFile(rulesFile, file, 0644)

}



func main() {
	fmt.Println()
	// kodel cia naudoji WithTimeout? Tavo programa nedirbs ilgiau kaip 5 min.
	// ar tikrai ta nori pasiekti cia?

	//M: 
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	wg.Add(1)

	// tokios funkcijos kurios viduje turi tickerius, turetu but paleidziamos
	// atskiroje rutinoje kaip go doEvery(..).

	//M: dar su lygiagrečiu programavimu truputį sunku, nes dabar infine laiką programa bėgs
	go doEvery(ctx, 30*time.Second, readFile)
	wg.Wait()

}
