package main

import (
	"fmt"
	"encoding/json"
	"os"
	"io/ioutil"
)

type RuleStruct struct{
	RulesArray []Rule `json:"rules"` 
}

type Rule struct{
	Crypto_id string `json:"crypto_id"`
	Price int `json:"price"`
	Rule string `json:"rule"`
}
func main () {
	fmt.Println()
	file, err := os.Open("rules.json")
	if err != nil{
		fmt.Println(err)
	}
//	var greater,lesser bool= false,false
	defer file.Close()

	byteValue, _ :=ioutil.ReadAll(file)

	var rules RuleStruct
	json.Unmarshal(byteValue,&rules)

	
	
}

