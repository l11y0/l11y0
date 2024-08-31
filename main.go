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
	// 切換到專案目錄
	err := os.Chdir(projectPath)
	if err != nil {
		log.Fatalf("無法切換到專案目錄：%v", err)
	}

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

	fmt.Println(mdContent)
	err = createWriteFile("README.md", mdContent)
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
		"git commit -m \"更新 README.md\"",
		"git push origin master",
	}

	for _, cmd := range commands {
		parts := strings.Fields(cmd)
		command := exec.Command(parts[0], parts[1:]...)

		// 設置環境變數以處理可能的憑證請求
		command.Env = append(os.Environ(),
			"GIT_ASKPASS=git-gui--askpass",
			"SSH_ASKPASS=git-gui--askpass",
		)

		output, err := command.CombinedOutput()
		if err != nil {
			log.Printf("執行命令 '%s' 時發生錯誤：%v", cmd, err)
			log.Printf("命令輸出：%s", output)
			return fmt.Errorf("執行命令 '%s' 失敗: %w", cmd, err)
		}
		fmt.Printf("命令 '%s' 的輸出：%s\n", cmd, output)
	}
	return nil
}

func createWriteFile(filename, content string) error {
	return ioutil.WriteFile(filename, []byte(content), 0644)
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
		return 0, 0, 0, fmt.Errorf("API 請求失敷鐨，狀態碼：%d，回應：%s", resp.StatusCode, string(body))
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

	return easy, medium, hard, nil
}

func readFile(filename string) (string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("讀取檔案 %s 時發生錯誤：%w", filename, err)
	}
	return string(data), nil
}
