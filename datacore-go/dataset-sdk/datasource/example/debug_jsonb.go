package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/example/datasource"
)

func isJSONString(s string) bool {
	s = strings.TrimSpace(s)
	return (strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")) || (strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]"))
}

func isBase64String(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return false
	}
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

func main() {
	base64Str := "eyJpZCI6ICIiLCAibmFtZSI6ICLmlrDmtYHnqIsiLCAibm9kZXMiOiBbeyJpZCI6ICJzdGFydF8xNzY3Nzc2Mjg0MzA0IiwgInR5cGUiOiAic3RhcnQiLCAic291cmNlcyI6IG51bGwsICJwb3NpdGlvbiI6IHsieCI6IDE1NS40OTgzLCAieSI6IDI1MS4yMjgyNH0sICJwcm9wZXJ0aWVzIjoge319LCB7ImlkIjogImh0dHBfMTc2Nzc3NjI4Nzg4MyIsICJ0eXBlIjogImh0dHAiLCAic291cmNlcyI6IFsic3RhcnRfMTc2Nzc3NjI4NDMwNCJdLCAicG9zaXRpb24iOiB7IngiOiA0NDYuNzAyNTgsICJ5IjogMjcxLjQwNjc0fSwgInByb3BlcnRpZXMiOiB7InVybCI6ICJodHRwOi8veG16YWlsLmNvbS9hdXRvU2V0L0NDQU0vYXV0by9nZXRBdXRvU2V0P2F1dG9Gb3JtSUQ9MjAxJnZhcmlhbnQ9MiIsICJtZXRob2QiOiAiR0VUIiwgImhlYWRlcnMiOiBbXSwgImNvbnRlbnRUeXBlIjogIiJ9fSwgeyJpZCI6ICJlbmRfMTc2Nzc3NjMwMDczOCIsICJ0eXBlIjogImVuZCIsICJzb3VyY2VzIjogWyJodHRwXzE3Njc3NzYyODc4ODMiXSwgInBvc2l0aW9uIjogeyJ4IjogNzQ2Ljk2NTcsICJ5IjogMjU4LjI2NTU2fSwgInByb3BlcnRpZXMiOiB7fX1dLCAibm9kZUNvdW50IjogMywgImNyZWF0ZWRfYXQiOiAiIiwgInN0YXJ0X25vZGUiOiAic3RhcnRfMTc2Nzc3NjI4NDMwNCIsICJ1cGRhdGVkX2F0IjogIiIsICJkZXNjcmlwdGlvbiI6ICIifQ=="

	fmt.Printf("Testing string: %s\n\n", base64Str[:50]+"...")

	fmt.Printf("Is JSON string? %v\n", isJSONString(base64Str))
	fmt.Printf("Is Base64 string? %v\n\n", isBase64String(base64Str))

	decoded, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		fmt.Printf("Decode error: %v\n", err)
	} else {
		fmt.Printf("Decoded: %s\n\n", string(decoded))
		fmt.Printf("Is decoded JSON? %v\n", isJSONString(string(decoded)))
	}

	fmt.Println("\n---\nNow let's test with datasource manager...")

	manager := datasource.GetManager()
	configJSON := `{"uuid":"post4bc7-9a41-4332-93a1-a60c4d8a7e19","type":"postgres","config":{"host":"121.43.142.153","port":5432,"database":"carrier","username":"carrier","password":"GNerfiSP4dpZjwcJ"}}`

	err = manager.AddDataSource(configJSON)
	if err != nil {
		fmt.Printf("Error adding datasource: %v\n", err)
		return
	}

	ds, err := manager.GetDataSource("post4bc7-9a41-4332-93a1-a60c4d8a7e19")
	if err != nil {
		fmt.Printf("Error getting datasource: %v\n", err)
		return
	}

	result, err := ds.GetAutoSet("workflow1", "workflow")
	if err != nil {
		fmt.Printf("Error getting autoset: %v\n", err)
		return
	}

	fmt.Printf("\nResult: %#v\n", result)

	if autoset, ok := result["autoset"]; ok {
		fmt.Printf("\nAutoset type: %T\n", autoset)
		fmt.Printf("Autoset value: %#v\n", autoset)
	}
}
