package main

import (
	"fmt"
	"encoding/json"
	"os"
	"io/ioutil"
	"time"
	"context"
	"net/http"
	"strconv"
)
const url ="https://api.coinlore.net/api/ticker/?id="
const rulesFile = "./assets/rules.json"
type RuleStruct struct{
	RulesArray []Rule `json:"rules"` 
}

type Rule struct{
	Crypto_id string `json:"crypto_id"`
	Price string `json:"price"`
	Rule string `json:"rule"`
}

type APIResult struct {
	Id string `json:"id"`
	Symbol string `json:"symbol"`
	Name string `json:"name"`
	NameId string `json:"nameid"`
	Rank int `json:"rank"`
	Price_usd string `json:"price_usd"`
	Percent_change24 string `json:"percent_change_24h"`
	Percent_change1 string `json:"percent_change_1h"`
	Percent_change7d string `json:"percent_change_7d"`
	Market_cap string `json:"market_cap_usd"`
	Volume24 string `json:"volume24"`
	Volume24_native string `json:"volume24_native"`
	Csupply string `json:"csupply"`
	Price_btc string `json:"price_btc"`
	Tsupply string `json:"tsupply"`
	Msupply string `json:"msupply"`

}

type shortAPI struct {
	Id string
	Name string
	Price string
}

func doEvery(cText context.Context, d time.Duration, f func(string,RuleStruct, time.Time)RuleStruct)error {
	ticker := time.Tick(1)
	var structure RuleStruct
	for {
		select{
		case <- cText.Done():
				return cText.Err()
		case x := <- ticker:
			ticker=time.Tick(d)
			structure  :=f(rulesFile,structure,x)
			for i:=0; i< len(structure.RulesArray);i++{
				fmt.Printf("Crypto_id:%v\n",structure.RulesArray[i].Crypto_id)
			}
			structure = checkRules(structure)
			reWritefile(structure)
		}
	
		
	}
}


func ReadFile(fileName string,rules RuleStruct, t time.Time)RuleStruct {
	file, err := os.Open(fileName)
	if err != nil{
		fmt.Println(err)
	}

	defer file.Close()
	fmt.Printf("File successfully opened at %v\n",t)

	byteValue, _ :=ioutil.ReadAll(file)


	json.Unmarshal(byteValue,&rules)
	return rules
}

func checkRules(structure RuleStruct)RuleStruct{

	
	array := structure.RulesArray
	for i:=0; i<len(array);i++{
		isUsed := false
		result:= GetAPI(array[i].Crypto_id)
		isUsed = findAPIresult(result, array[i],isUsed)
		if isUsed {
			array[0],array[i] = array[i], array[0]
			structure.RulesArray = array[1:]

		}
	}
	return structure
}

func findAPIresult(API shortAPI, rule Rule, IsUsed bool)bool{
	floatRulePrice, _ := strconv.ParseFloat(rule.Price,64)
	floatApiPrice, _ := strconv.ParseFloat(API.Price,64)
	switch rule.Rule {
		case "lt":
			if (floatApiPrice < floatRulePrice && !IsUsed) {
				fmt.Printf("Cryptocurrency id:"+ API.Id + " "+ API.Name + " is lower than %v\n",rule.Price)
				IsUsed = true
			}
		case "gt":
			if (floatApiPrice > floatRulePrice && !IsUsed) {
				fmt.Printf("Cryptocurrency id:"+ API.Id + " "+ API.Name + " is grater than %v\n", rule.Price)
				IsUsed = true
			}
	
	}
	return IsUsed

}

func GetAPI(Id string) shortAPI {

	URL := fmt.Sprintf(url+Id)
	response, err := http.Get(URL)
    if err != nil {
        fmt.Print(err.Error())
        os.Exit(1)
    }

	responseData, _ := ioutil.ReadAll(response.Body)
	
	//responseData[:len(responseData)] =125
	//trimedResponseData :=responseData[0:]
	//trimedResponseData2:=trimedResponseData[:len(trimedResponseData)-1]
	

	//var responseObject APIStruct
	var responseObject []APIResult
	error:= json.Unmarshal(responseData, &responseObject)
	
	if error != nil {
		fmt.Printf("There was an error decoding the json. err = %s", error)
	}
//	oneAPI :=responseObject.ApiArray[0]
//	result := createShortApi(oneAPI.Id,oneAPI.Name,oneAPI.Price_usd)
	result := createShortApi(responseObject[0].Id,responseObject[0].Name,responseObject[0].Price_usd)
	return result
}

func reWritefile(structure RuleStruct){
	file, _ := json.MarshalIndent(structure, "", " ")
	_ =ioutil.WriteFile("./assets/rules.json",file, 0644)

}

func createShortApi(id string,name string, price string) shortAPI {
	API := shortAPI {Id: id, Name: name, Price: price}
	return API
}


func main () {
	fmt.Println()
	cText, cancel := context.WithTimeout(context.Background(), 5* time.Minute)
	defer cancel()
	doEvery(cText,30*time.Second,ReadFile)
	

	
	
}

