package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var userName = "l11y0" // 請確保這是您的 LeetCode 用戶名

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
		log.Fatalf("寫入檔案時發生錯誤：%v", err)
	}
	err = updateGithub()
	if err != nil {
		log.Fatalf("更新 GitHub 時發生錯誤：%v", err)
	}
}

func checkFileIsExist(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func updateGithub() error {
	commands := []string{
		"git add README.md",
		"git commit -m \"update\"",
		"git push",
	}

	for _, cmd := range commands {
		parts := strings.Fields(cmd)
		command := exec.Command(parts[0], parts[1:]...)
		command.Dir = filepath.Dir(os.Args[0]) // 設置工作目錄為程式所在的目錄
		output, err := command.CombinedOutput()
		if err != nil {
			return fmt.Errorf("執行命令 '%s' 時發生錯誤：%v，輸出：%s", cmd, err, output)
		}
		fmt.Printf("命令 '%s' 的輸出：%s\n", cmd, output)
	}
	return nil
}

func createWriteFile(mdContent string) error {
	return ioutil.WriteFile("README.md", []byte(mdContent), 0644)
}

func getQuestionProgressInfo() (easy, medium, hard int) {
	client := &http.Client{}
	jsonStr := `{
		"query": "query userProfileUserQuestionProgressV2($username: String!) { userProfileUserQuestionProgressV2(username: $username) { numAcceptedQuestions { difficulty count } } }",
		"variables": {
			"username": "` + userName + `"
		}
	}`
	req, err := http.NewRequest("POST", "https://leetcode.com/graphql/", strings.NewReader(jsonStr))
	if err != nil {
		log.Fatalf("創建請求時發生錯誤：%v", err)
	}
	req.Header.Add("content-type", "application/json")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("發送請求時發生錯誤：%v", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("API 請求失敗，狀態碼：%d，回應：%s", resp.StatusCode, string(body))
	}

	fmt.Println("API 回應：", string(body))

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
		log.Fatalf("解析 JSON 時發生錯誤：%v", err)
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
		log.Fatalf("讀取模板檔案時發生錯誤：%v", err)
	}
	return string(data)
}
