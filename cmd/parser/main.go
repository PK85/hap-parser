package main

import (
	"fmt"
	"hap-parser/internal/rule"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

const (
	RULE_OK1 = "aws"
	RULE_OK2 = "aws -> EU, S"
	RULE_OK3 = "aws(PR=cf-eu11) -> EU"
	RULE_OK4 = "aws(PR=cf-eu11) ->	 EU"

	RULE_OUT_NOT_OK1 = "aws(PR=cf-eu11) -> S -> EU"  // "Separator -> is optional and max 1 occurance. Current count: 2"
	RULE_OUT_NOT_OK2 = "aws(PR=cf-eu11) ->,S"        // "Output Attributes cannot be empty. Attribute No. 1 is empty"
	RULE_OUT_NOT_OK3 = "aws(PR=cf-eu11) -> ,,"       // "Output Attributes cannot be empty. Attribute No. 1 is empty"
	RULE_OUT_NOT_OK4 = "aws(PR=cf-eu11) -> S,"       // "Output Attributes cannot be empty. Attribute No. 2 is empty"
	RULE_OUT_NOT_OK5 = "aws(PR=cf-eu11) -> S,AA"     // "Allowed Output Attributes are [S EU]. Attribute No. 2 has not supported value: AA"
	RULE_OUT_NOT_OK6 = "aws(PR=cf-eu11) ->AA"        // "Allowed Output Attributes are [S EU]. Attribute No. 1 has not supported value: AA"
	RULE_OUT_NOT_OK7 = "aws(PR=cf-eu11) -> EU, S, S" // "Max number of Output Attributes equals number of supported Output Attributes. Supported: 2, current: 3"
	RULE_OUT_NOT_OK8 = "aws(PR=cf-eu11) -> EU, EU"   // "Allowed Output Attributes could be specified once. Attribute No. 2 is duplicated, value: EU"
	RULE_OUT_NOT_OK9 = "aws(PR=cf-eu11) -> S, S"     // "Allowed Output Attributes could be specified once. Attribute No. 2 is duplicated, value: S"}

	RULE_IN_NOT_OK1  = "aaws"                           // "Allowed Plans: [aws azure gcp]. Plan and/or Input Attributes has incorrect syntax: aaws"
	RULE_IN_NOT_OK2  = "awss"                           // "Allowed Plans: [aws azure gcp]. Plan and/or Input Attributes has incorrect syntax: awss"
	RULE_IN_NOT_OK3  = "aws()"                          // "Allowed Plans: [aws azure gcp]. Plan and/or Input Attributes has incorrect syntax: aws()"
	RULE_IN_NOT_OK4  = "aws("                           // "Allowed Plans: [aws azure gcp]. Plan and/or Input Attributes has incorrect syntax: aws("
	RULE_IN_NOT_OK5  = "aws(PR=cf-eu11"                 // "Allowed Plans: [aws azure gcp]. Plan and/or Input Attributes has incorrect syntax: aws(PR=cf-eu11"
	RULE_IN_NOT_OK6  = "awsPR=cf-eu11)"                 // "Allowed Plans: [aws azure gcp]. Plan and/or Input Attributes has incorrect syntax: awsPR=cf-eu11)"
	RULE_IN_NOT_OK7  = "aws(X)"                         // "Input Attributes must contain `=` character. Attribute No. 1 is invalid, value: X"
	RULE_IN_NOT_OK8  = "aws(, )"                        // "Input Attributes cannot be empty. Attribute No. 1 is empty"
	RULE_IN_NOT_OK9  = "aws(PR)"                        // "Input Attributes must contain `=` character.Attribute No. 1 is invalid, value: PR"
	RULE_IN_NOT_OK10 = "aws(PR=cf-eu11, PR) "           // "Input Attributes must contain `=` character. Attribute No. 2 is ivalid, value: PR"
	RULE_IN_NOT_OK11 = "aws(PR=cf-=eu11)"               // "Input Attributes must contain only one `=` character. Attribute No. 1 is invalid, value: PR=cf-=eu11"
	RULE_IN_NOT_OK12 = "aws(PR=cf.PR=eu11)"             // "Input Attributes must contain only one `=` character. Attribute No. 1 is invalid, value: PR=cf.PR=eu11"
	RULE_IN_NOT_OK13 = "aws(HR=westeu, PR=cf-=eu11)"    // "Input Attributes must contain only one `=` character. Attribute No. 2 is invalid, value: PR=cf-=eu11"}
	RULE_IN_NOT_OK14 = "aws(PR=cf-eu11, PR=)"           // "Input Attributes could be specified once. Attribute No. 2 is duplicated, value: PR"
	RULE_IN_NOT_OK15 = "aws( PR=)"                      // "Input Attribute PR cannot be empty. Attribute No. 1, value: PR="
	RULE_IN_NOT_OK16 = "aws(PR=cf-eu11, PR=cf-eu11)"    // "Input Attributes could be specified once. Attribute No. 2 is duplicated, value: PR"
	RULE_IN_NOT_OK17 = "aws(HR=cf-eu11, PR=)"           // "Input Attribute PR cannot be empty. Attribute No. 2, value: PR="}
	RULE_IN_NOT_OK18 = "aws( HR=)"                      // "Input Attribute HR cannot be empty. Attribute No. 1, value: HR="
	RULE_IN_NOT_OK19 = "aws(HR=cf-eu11, HR=cf-eu11)"    // "Input Attributes could be specified once. Attribute No. 2 is duplicated, value: HR"
	RULE_IN_NOT_OK20 = "aws(PR=cf-eu11, HR=12, PR=88) " // "Max number of Input Attributes equals number of supported Input Attributes. Supported: 2, current: 3"
	RULE_IN_NOT_OK21 = "aws(KK=cf-eu11) "               // "Allowed Input Attributes are [PR HR]. Attribute No. 1 has not supported value: KK"
	RULE_IN_NOT_OK22 = "aws(PR=cf-eu11, KK=cf-eu11) "   // "Allowed Input Attributes are [PR HR]. Attribute No. 2 has not supported value: KK"
	RULE_IN_NOT_OK23 = "aws(KK=cf-eu11, PR=cf-eu11 ) "  // "Allowed Input Attributes are [PR HR]. Attribute No. 1 has not supported value: KK"
)

func main() {
	var c conf
	for index, ruleRaw := range c.getConf().Rules {
		rule := rule.ConvertToRule(ruleRaw)
		if !rule.IsInvalid {
			fmt.Printf("Rule No. %d is OK, value: [%s]	- searchLabels: %s\n", index+1, rule.Raw, *rule.SearchLabels)
		} else {
			fmt.Printf("Rule No. %d is Invalid, value: [%s]	- reason: %s\n\n", index+1, rule.Raw, rule.InvalidMessage)
		}
	}
}

type conf struct {
	Rules []string `yaml:"rule"`
}

func (c *conf) getConf() *conf {

	yamlFile, err := os.ReadFile("resources/rules.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c
}
