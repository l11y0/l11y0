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

var userSlug = "l11y0"
var projectPath = filepath.Join("C:\\Users", "黃韻如", "Documents", "l11y0")

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// 切換到專案目錄
	err := os.Chdir(projectPath)
	if err != nil {
		log.Fatalf("無法切換到專案目錄：%v", err)
	}
	log.Printf("當前工作目錄: %s", projectPath)

	easy, medium, hard, err := getQuestionProgressInfo()
	if err != nil {
		log.Fatalf("獲取題目進度信息時發生錯誤：%v", err)
	}

	mdContent, err := readFile("README-TEMP.md")
	if err != nil {
		log.Fatalf("讀取 README-TEMP.md 時發生錯誤：%v", err)
	}

	totalSolved := easy + medium + hard
	mdContent = strings.ReplaceAll(mdContent, `[[1]]`, strconv.Itoa(totalSolved))
	mdContent = strings.ReplaceAll(mdContent, `[[2]]`, strconv.Itoa(easy))
	mdContent = strings.ReplaceAll(mdContent, `[[3]]`, strconv.Itoa(medium))
	mdContent = strings.ReplaceAll(mdContent, `[[4]]`, strconv.Itoa(hard))

	log.Println("更新後的 README 內容:")
	log.Println(mdContent)

	err = createWriteFile("README.md", mdContent)
	if err != nil {
		log.Fatalf("寫入檔案時發生錯誤：%v", err)
	}
	log.Println("README.md 已成功更新")

	err = updateGithub()
	if err != nil {
		log.Fatalf("更新 GitHub 時發生錯誤：%v", err)
	}
	log.Println("GitHub 更新成功")
}

func updateGithub() error {
	commands := [][]string{
		{"git", "add", "README.md"},
		{"git", "commit", "-m", "更新 README.md"},
		{"git", "push", "origin", "master"},
	}

	for _, cmd := range commands {
		log.Printf("執行命令: %s", strings.Join(cmd, " "))
		command := exec.Command(cmd[0], cmd[1:]...)

		output, err := command.CombinedOutput()
		log.Printf("命令輸出:\n%s", string(output))

		if err != nil {
			return fmt.Errorf("執行命令 '%s' 失敗: %w\n輸出: %s", strings.Join(cmd, " "), err, string(output))
		}
	}
	return nil
}

func createWriteFile(filename, content string) error {
	err := ioutil.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("寫入檔案 %s 時發生錯誤: %w", filename, err)
	}
	log.Printf("成功寫入檔案: %s", filename)
	return nil
}

func getQuestionProgressInfo() (easy, medium, hard int, err error) {
	client := &http.Client{}
	jsonStr := `{
		"query": "query userProfileUserQuestionProgressV2($userSlug: String!) { userProfileUserQuestionProgressV2(userSlug: $userSlug) { numAcceptedQuestions { difficulty count } } }",
		"variables": {
			"userSlug": "` + userSlug + `"
		}
	}`
	req, err := http.NewRequest("POST", "https://leetcode.com/graphql/", strings.NewReader(jsonStr))
	if err != nil {
		return 0, 0, 0, fmt.Errorf("創建請求時發生錯誤：%w", err)
	}
	req.Header.Add("content-type", "application/json")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("發送請求時發生錯誤：%w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("讀取回應內容時發生錯誤：%w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return 0, 0, 0, fmt.Errorf("API 請求失敗，狀態碼：%d，回應：%s", resp.StatusCode, string(body))
	}

	log.Println("API 回應：", string(body))

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
		return 0, 0, 0, fmt.Errorf("解析 JSON 時發生錯誤：%w", err)
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

	log.Printf("題目統計: Easy: %d, Medium: %d, Hard: %d", easy, medium, hard)
	return easy, medium, hard, nil
}

func readFile(filename string) (string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("讀取檔案 %s 時發生錯誤：%w", filename, err)
	}
	log.Printf("成功讀取檔案: %s", filename)
	return string(data), nil
}
