// This file was automatically generated by "microgen 0.6.0" utility.
// Please, do not edit.
package stringsvc

import entity "github.com/devimteam/microgen/example/svc/entity"

type UppercaseRequest struct {
	Str []map[string]interface{} `json:"str"` // This field was defined with ellipsis (...).
}

type UppercaseResponse struct {
	Ans string `json:"ans"`
}

type CountRequest struct {
	Text   string `json:"text"`
	Symbol string `json:"symbol"`
}

type CountResponse struct {
	Count     int   `json:"count"`
	Positions []int `json:"positions"`
}

type TestCaseRequest struct {
	Comments []*entity.Comment `json:"comments"`
}

type TestCaseResponse struct {
	Tree map[string]int `json:"tree"`
}
