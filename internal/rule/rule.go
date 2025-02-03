package rule

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	INPUT_LEFT_SEP  = "("
	INPUT_RIGHT_SEP = ")"
	OUTPUT_SEP      = "->"
)

var AllowedPlans = []string{"aws", "azure", "gcp", "preview", "azure_lite", "free", "trial", "sap-converged-cloud"}
var AllowedInputAttrs = []string{"PR", "HR"}
var AllowedOutputAttrs = []string{"S", "EU"}

type Rule struct {
	Raw               string  // raw rule, example: aws(PR=cf-eu11) -> EU
	Plan              string  // always mandatory, must be allowed plan
	PlatformRegion    *string // if nil do not use in SearchLabels, allowed values: "*", example: "cf-eu11", not allowed empty string
	HyperscalerRegion *string // if nil do not use in SearchLabels, allowed values: "*", example: "westeu", not allowed empty string
	EuAccess          bool    // if false do not use in SearchLabels
	Shared            bool    // if false do not use in SearchLabels
	IsInvalid         bool    // if false check InvalidMessage for details
	InvalidMessage    string  // empty if rule is ok
	SearchLabels      *string // if nil rule is invalid
}

func (r Rule) String() string {
	return fmt.Sprintf("%#v", r)
}

/*
hyperscalerType: plan_PR_HR
euAccess: true/false
shared: true/false
*/
func (r *Rule) calculateSearchLabels() {
	var result string
	result = "hyperscalerType: " + getHyperscalerName(r.Plan)
	if r.PlatformRegion != nil {
		result += "_" + *r.PlatformRegion
	}
	if r.HyperscalerRegion != nil {
		result += "_" + *r.HyperscalerRegion
	}
	if r.EuAccess {
		result += "; " + "euAccess: true"
	}
	if r.Shared {
		result += "; " + "shared: true"
	}
	r.SearchLabels = &result
}

func ConvertToRule(input string) (result Rule) {
	result.Raw = input
	modRule := removeWhiteChars(input)
	// fmt.Println(modRule)

	// OUTPUT
	count := countMatches(modRule, OUTPUT_SEP) // 0 or 1 allowed
	if count > 1 {
		result.IsInvalid = true
		result.InvalidMessage = fmt.Sprintf("Separator %s is optional and max 1 occurance. Current count: %d", OUTPUT_SEP, count)
		return
	}

	inputLeft, outputRight, isSepFound := strings.Cut(modRule, "->")
	if isSepFound {
		tempRight := strings.Split(outputRight, ",")
		if len(tempRight) > len(AllowedOutputAttrs) {
			result.IsInvalid = true
			result.InvalidMessage = fmt.Sprintf("Max number of Output Attributes equals number of supported Output Attributes. "+
				"Supported: %d, current: %d", len(AllowedOutputAttrs), len(tempRight))
			return
		}
		// fmt.Println(len(tempRight))
		for index, value := range tempRight {
			if len(value) == 0 {
				result.IsInvalid = true
				result.InvalidMessage = fmt.Sprintf("Output Attributes cannot be empty. Attribute No. %d is empty", index+1)
				return
			}
			if countOutAttrsMatches(value) == 0 {
				result.IsInvalid = true
				result.InvalidMessage = fmt.Sprintf("Allowed Output Attributes are %v. Attribute No. %d has not supported value: %s",
					AllowedOutputAttrs, index+1, value)
				return
			}
			if value == "EU" {
				if result.EuAccess {
					result.IsInvalid = true
					result.InvalidMessage = fmt.Sprintf("Allowed Output Attributes could be specified once. Attribute No. %d is duplicated, value: %s",
						index+1, value)
					return
				} else {
					result.EuAccess = true
				}
			}
			if value == "S" {
				if result.Shared {
					result.IsInvalid = true
					result.InvalidMessage = fmt.Sprintf("Allowed Output Attributes could be specified once. Attribute No. %d is duplicated, value: %s",
						index+1, value)
					return
				} else {
					result.Shared = true
				}
			}
		}
	}

	// INPUT
	var isAllowed bool
	var allowedPlan string
	var allowedRight string

	for _, plan := range AllowedPlans {
		left, right, isPlanFound := strings.Cut(inputLeft, plan)
		if isPlanFound && len(left) == 0 && (len(right) == 0 || (strings.HasPrefix(right, "(") && strings.HasSuffix(right, ")") && len(right) != 2)) {
			// ok
			isAllowed = true
			allowedPlan = plan
			allowedRight = right
		}
	}
	if isAllowed {
		result.Plan = allowedPlan
		if len(allowedRight) != 0 { // when true it has at least len(3) "(X)"
			tempRightStr := allowedRight[1 : len(allowedRight)-1] //remove "("" and ")"
			// fmt.Println(tempRightStr)
			// only PR=value and HR=value allowed, once. All other combination invalid, like duplication, wrong names, missing =, missing values etc
			tempRights := strings.Split(tempRightStr, ",")
			if len(tempRights) > len(AllowedInputAttrs) {
				result.IsInvalid = true
				result.InvalidMessage = fmt.Sprintf("Max number of Input Attributes equals number of supported Input Attributes. "+
					"Supported: %d, current: %d", len(AllowedInputAttrs), len(tempRights))
				return
			}

			for index, inputAttr := range tempRights {
				if len(inputAttr) == 0 {
					result.IsInvalid = true
					result.InvalidMessage = fmt.Sprintf("Input Attributes cannot be empty. Attribute No. %d is empty", index+1)
					return
				}
				inputAttrArray := strings.Split(inputAttr, "=")
				if len(inputAttrArray) == 1 { // "=" no found, fail
					result.IsInvalid = true
					result.InvalidMessage = fmt.Sprintf("Input Attributes must contain `=` character. "+
						"Attribute No. %d is invalid, value: %s", index+1, inputAttr)
					return
				}
				if len(inputAttrArray) > 2 { // many "=" found, fail
					result.IsInvalid = true
					result.InvalidMessage = fmt.Sprintf("Input Attributes must contain only one `=` character. "+
						"Attribute No. %d is invalid, value: %s", index+1, inputAttr)
					return
				}
				if countInAttrsMatches(inputAttrArray[0]) == 0 {
					result.IsInvalid = true
					result.InvalidMessage = fmt.Sprintf("Allowed Input Attributes are %v. Attribute No. %d has not supported value: %s",
						AllowedInputAttrs, index+1, inputAttrArray[0])
					return
				}
				if inputAttrArray[0] == "PR" {
					if result.PlatformRegion != nil {
						result.IsInvalid = true
						result.InvalidMessage = fmt.Sprintf("Input Attributes could be specified once. Attribute No. %d is duplicated, value: %s",
							index+1, "PR")
						return
					} else {
						if len(inputAttrArray[1]) == 0 {
							result.IsInvalid = true
							result.InvalidMessage = fmt.Sprintf("Input Attribute PR cannot be empty. Attribute No. %d, value: PR=%s",
								index+1, inputAttrArray[1])
							return
						} else {
							result.PlatformRegion = &inputAttrArray[1]
						}
					}
				}

				if inputAttrArray[0] == "HR" {
					if result.HyperscalerRegion != nil {
						result.IsInvalid = true
						result.InvalidMessage = fmt.Sprintf("Input Attributes could be specified once. Attribute No. %d is duplicated, value: %s",
							index+1, "HR")
						return
					} else {
						if len(inputAttrArray[1]) == 0 {
							result.IsInvalid = true
							result.InvalidMessage = fmt.Sprintf("Input Attribute HR cannot be empty. Attribute No. %d, value: HR=%s",
								index+1, inputAttrArray[1])
							return
						} else {
							result.HyperscalerRegion = &inputAttrArray[1]
						}
					}
				}
			}
		}
	} else {
		result.IsInvalid = true
		result.InvalidMessage = fmt.Sprintf("Allowed Plans: %v. Plan and/or Input Attributes has incorrect syntax: %s",
			AllowedPlans, inputLeft)
		return
	}

	result.calculateSearchLabels()
	return result
}

func removeWhiteChars(str string) string {
	return strings.Join(strings.Fields(str), "")
}

func countMatches(input string, sep string) int {
	r := regexp.MustCompile(sep)
	matches := r.FindAllStringIndex(input, -1)
	return len(matches)
}

func countOutAttrsMatches(input string) int {
	var str string
	for i, attr := range AllowedOutputAttrs {
		str += attr
		if len(AllowedOutputAttrs) != i+1 { //last separator not added
			str += "|"
		}
	}

	r := regexp.MustCompile(str)
	matches := r.FindAllStringIndex(input, -1)
	return len(matches)
}

func countInAttrsMatches(input string) int {
	var str string
	for i, attr := range AllowedInputAttrs {
		str += attr
		if len(AllowedOutputAttrs) != i+1 { //last separator not added
			str += "|"
		}
	}

	r := regexp.MustCompile(str)
	matches := r.FindAllStringIndex(input, -1)
	return len(matches)
}

func getHyperscalerName(plan string) (result string) {
	if plan == "aws" || plan == "gcp" || plan == "azure" || plan == "azure_lite" {
		return plan
	} else if plan == "trial" {
		return "aws"
	} else if plan == "free" {
		return "aws/azure"
	} else if plan == "sap-converged-cloud" {
		return "openstack"
	} else if plan == "preview" {
		return "aws"
	} else {
		return ""
	}
}
