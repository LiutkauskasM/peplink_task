package main

import (
	"fmt"
	"encoding/json"
	"os"
	"io/ioutil"
	"time"
	"context"
)

func doEvery(cText context.Context, d time.Duration, f func(string,RuleStruct, time.Time)RuleStruct)error {
	ticker := time.Tick(1)
	file  := "rules.json"
	var structure RuleStruct
	for {
		select{
		case <- cText.Done():
				return cText.Err()
		case x := <- ticker:
			ticker=time.Tick(d)
			structure  :=f(file,structure,x)
			fmt.Printf("Crypto_id:%v\n",structure.RulesArray[0].Crypto_id)
		
		}
	
		
	}
}


func ReadFile(fileName string,rules RuleStruct, t time.Time)RuleStruct {
	file, err := os.Open(fileName)
	if err != nil{
		fmt.Println(err)
	}
//	var greater,lesser bool= false,false
	defer file.Close()
	fmt.Printf("File successfully opened at %v\n",t)

	byteValue, _ :=ioutil.ReadAll(file)


	json.Unmarshal(byteValue,&rules)
	return rules
}


type RuleStruct struct{
	RulesArray []Rule `json:"rules"` 
}

type Rule struct{
	Crypto_id string `json:"crypto_id"`
	Price int `json:"price"`
	Rule string `json:"rule"`
	isUsed bool
}
func main () {
	fmt.Println()
	cText, cancel := context.WithTimeout(context.Background(), 1* time.Minute)
	defer cancel()
	doEvery(cText,30*time.Second,ReadFile)
	

	
	
}

