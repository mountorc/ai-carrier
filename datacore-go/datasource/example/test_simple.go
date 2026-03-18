package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
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

	fmt.Println("=== Testing Simple Functions ===")
	fmt.Printf("Is JSON: %v\n", isJSONString(base64Str))
	fmt.Printf("Is Base64: %v\n", isBase64String(base64Str))

	if isBase64String(base64Str) {
		fmt.Println("\nDecoding Base64...")
		decoded, err := base64.StdEncoding.DecodeString(base64Str)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			decodedStr := string(decoded)
			fmt.Printf("Decoded: %s...\n", decodedStr[:100])
			fmt.Printf("Is decoded JSON: %v\n", isJSONString(decodedStr))

			if isJSONString(decodedStr) {
				var jsonVal interface{}
				if err := json.Unmarshal(decoded, &jsonVal); err == nil {
					fmt.Printf("✓ Successfully unmarshaled JSON: %#v\n", jsonVal)
				} else {
					fmt.Printf("✗ Unmarshal error: %v\n", err)
				}
			}
		}
	}
}
