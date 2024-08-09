package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var userSlug = "l11y0" // 請確保這是您的 LeetCode 用戶名

func main() {
	easy, medium, hard := getQuestionProgressInfo()
	mdContent := readFile()
	mdContent = strings.ReplaceAll(mdContent, `[[1]]`, strconv.Itoa(easy+medium+hard))
	mdContent = strings.ReplaceAll(mdContent, `[[2]]`, strconv.Itoa(easy))
	mdContent = strings.ReplaceAll(mdContent, `[[3]]`, strconv.Itoa(medium))
	mdContent = strings.ReplaceAll(mdContent, `[[4]]`, strconv.Itoa(hard))
	fmt.Println(mdContent)
	err := createWriteFile(mdContent)
	if err != nil {
		log.Fatalf("Error writing file: %v", err)
	}
	err = updateGithub()
	if err != nil {
		log.Fatalf("Error updating GitHub: %v", err)
	}
}

func checkFileIsExist(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func updateGithub() error {
	cmd := exec.Command("sh", "./auto.sh")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error executing script: %v, output: %s", err, output)
	}
	fmt.Println(string(output))
	return nil
}

func createWriteFile(mdContent string) error {
	return ioutil.WriteFile("README.md", []byte(mdContent), 0644)
}

func getQuestionProgressInfo() (easy, medium, hard int) {
	client := &http.Client{}
	jsonStr := `{
		"query": "query userProfileUserQuestionProgressV2($userSlug: String!) { userProfileUserQuestionProgressV2(userSlug: $userSlug) { numAcceptedQuestions { difficulty count } } }",
		"variables": {
			"userSlug": "` + userSlug + `"
		}
	}`
	req, err := http.NewRequest("POST", "https://leetcode.com/graphql/", strings.NewReader(jsonStr))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	req.Header.Add("content-type", "application/json")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("API request failed with status code: %d, response: %s", resp.StatusCode, string(body))
	}

	fmt.Println("API Response:", string(body))

	var response struct {
		Data struct {
			UserProfileUserQuestionProgressV2 struct {
				NumAcceptedQuestions []struct {
					Difficulty string `json:"difficulty"`
					Count      int    `json:"count"`
				} `json:"numAcceptedQuestions"`
			} `json:"userProfileUserQuestionProgressV2"`
		} `json:"data"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Fatalf("Error unmarshaling JSON: %v", err)
	}

	for _, item := range response.Data.UserProfileUserQuestionProgressV2.NumAcceptedQuestions {
		switch item.Difficulty {
		case "EASY":
			easy = item.Count
		case "MEDIUM":
			medium = item.Count
		case "HARD":
			hard = item.Count
		}
	}

	return
}

func readFile() string {
	data, err := ioutil.ReadFile("README-TEMP.md")
	if err != nil {
		log.Fatalf("Error reading template file: %v", err)
	}
	return string(data)
}
