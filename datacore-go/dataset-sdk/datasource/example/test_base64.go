package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

func main() {
	testStr := "eyJpZCI6ICIiLCAibmFtZSI6ICLmlrDmtYHnqIsiLCAibm9kZXMiOiBbeyJpZCI6ICJzdGFydF8xNzY3Nzc2Mjg0MzA0IiwgInR5cGUiOiAic3RhcnQiLCAic291cmNlcyI6IG51bGwsICJwb3NpdGlvbiI6IHsieCI6IDE1NS40OTgzLCAieSI6IDI1MS4yMjgyNH0sICJwcm9wZXJ0aWVzIjoge319LCB7ImlkIjogImh0dHBfMTc2Nzc3NjI4Nzg4MyIsICJ0eXBlIjogImh0dHAiLCAic291cmNlcyI6IFsic3RhcnRfMTc2Nzc3NjI4NDMwNCJdLCAicG9zaXRpb24iOiB7IngiOiA0NDYuNzAyNTgsICJ5IjogMjcxLjQwNjc0fSwgInByb3BlcnRpZXMiOiB7InVybCI6ICJodHRwOi8veG16YWlsLmNvbS9hdXRvU2V0L0NDQU0vYXV0by9nZXRBdXRvU2V0P2F1dG9Gb3JtSUQ9MjAxJnZhcmlhbnQ9MiIsICJtZXRob2QiOiAiR0VUIiwgImhlYWRlcnMiOiBbXSwgImNvbnRlbnRUeXBlIjogIiJ9fSwgeyJpZCI6ICJlbmRfMTc2Nzc3NjMwMDczOCIsICJ0eXBlIjogImVuZCIsICJzb3VyY2VzIjogWyJodHRwXzE3Njc3NzYyODc4ODMiXSwgInBvc2l0aW9uIjogeyJ4IjogNzQ2Ljk2NTcsICJ5IjogMjU4LjI2NTU2fSwgInByb3BlcnRpZXMiOiB7fX1dLCAibm9kZUNvdW50IjogMywgImNyZWF0ZWRfYXQiOiAiIiwgInN0YXJ0X25vZGUiOiAic3RhcnRfMTc2Nzc3NjI4NDMwNCIsICJ1cGRhdGVkX2F0IjogIiIsICJkZXNjcmlwdGlvbiI6ICIifQ=="

	fmt.Println("Testing Base64 decoding...")
	decoded, err := base64.StdEncoding.DecodeString(testStr)
	if err != nil {
		fmt.Printf("Decode error: %v\n", err)
		return
	}
	fmt.Printf("Decoded: %s\n", string(decoded))

	var jsonVal interface{}
	if err := json.Unmarshal(decoded, &jsonVal); err != nil {
		fmt.Printf("JSON unmarshal error: %v\n", err)
		return
	}
	fmt.Println("Successfully unmarshaled JSON!")
	jsonResult, _ := json.MarshalIndent(jsonVal, "", "  ")
	fmt.Println(string(jsonResult))
}
