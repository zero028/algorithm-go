package test

import (
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/go-resty/resty/v2"

	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

var m = map[string]string{
	"Easy":   "简单",
	"Medium": "中等",
	"Hard":   "困难",
}

const solutionContent = `package leetcode

%s
`

const solutionTestContent = `package leetcode

import "testing"

func Test(t *testing.T) {
	t.Log()
}
`

// https://cattiek.site/2019/03/03/Leetcode%E7%88%AC%E8%99%AB%E5%AE%9E%E8%B7%B5/#%E9%A2%98%E7%9B%AE%E8%AF%A6%E6%83%85-%E5%85%8D%E7%99%BB%E9%99%86%E5%8F%AF%E8%8E%B7%E5%8F%96
func getContent(slug string) QuestionContent {
	// 请求的 URL
	url := "https://leetcode.cn/graphql/"

	// 设置请求头
	headers := map[string]string{
		"Accept-Language": "zh-CN,zh;q=0.9",
		"Content-Type":    "application/json",
	}

	body := map[string]interface{}{
		"query": `query questionTranslations($titleSlug: String!) {
  question(titleSlug: $titleSlug) {
    translatedTitle
    translatedContent
    questionFrontendId
    difficulty
    titleSlug
    topicTags{
      translatedName
    }
    codeSnippets {
      lang
      langSlug
      code
      __typename
    }
  }
}`,
		"variables": map[string]string{
			"titleSlug": slug,
		},
		"operationName": "questionTranslations",
	}
	var resp QuestionContent

	resty.New().SetJSONMarshaler(json.Marshal).SetJSONUnmarshaler(json.Unmarshal).R().
		SetHeaders(headers).
		SetBody(body).
		SetResult(&resp).
		Post(url)

	converter := md.NewConverter("", true, nil)

	markdown, err := converter.ConvertString(resp.Data.Question.TranslatedContent)
	if err != nil {
		log.Fatal(err)
	}
	resp.Data.Question.TranslatedContent = markdown
	resp.Data.Question.Difficulty = m[resp.Data.Question.Difficulty]
	return resp
}

type QuestionContent struct {
	Data struct {
		Question struct {
			TranslatedTitle    string `json:"translatedTitle"`
			TranslatedContent  string `json:"translatedContent"`
			QuestionFrontendId string `json:"questionFrontendId"`
			Difficulty         string `json:"difficulty"`
			TitleSlug          string `json:"titleSlug"`
			TopicTags          []struct {
				Name           string `json:"name"`
				Slug           string `json:"slug"`
				TranslatedName string `json:"translatedName"`
			} `json:"topicTags"`
			CodeSnippets []struct {
				Lang     string `json:"lang"`
				LangSlug string `json:"langSlug"`
				Code     string `json:"code"`
				Typename string `json:"__typename"`
			} `json:"codeSnippets"`
		} `json:"question"`
	} `json:"data"`
}

func GenFile(questionContent QuestionContent) error {
	fileContent := GenContent(questionContent)
	f := fmt.Sprintf("../content/leetcode/%s%s.md", questionContent.Data.Question.QuestionFrontendId, questionContent.Data.Question.TranslatedTitle)
	create, err := os.Create(f)
	if err != nil {
		return err
	}
	_, err = create.WriteString(fileContent)
	cmd := exec.Command("git", "add", f)
	cmd.Run()
	return err
}

func GenContent(questionContent QuestionContent) string {
	tags := ""
	for index := range questionContent.Data.Question.TopicTags {
		tags += `
  - ` + questionContent.Data.Question.TopicTags[index].TranslatedName
	}

	return fmt.Sprintf(
		`---
title: %s
categories:
  - %s
tags: %s
slug: %s
number: %s
---

## 题目描述：

%s

---
## 解题分析及思路：

### 方法：方法

**思路：**


**复杂度：**

- 时间复杂度：O(N * M)
- 空间复杂度：O(1)

**执行结果：**

- 执行耗时:1 ms,击败了40.84 的Go用户
- 内存消耗:2.4 MB,击败了28.50 的Go用户
`, questionContent.Data.Question.TranslatedTitle,
		questionContent.Data.Question.Difficulty,
		tags,
		questionContent.Data.Question.TitleSlug,
		questionContent.Data.Question.QuestionFrontendId,
		questionContent.Data.Question.TranslatedContent)
}

func GenQuestionCode(resp QuestionContent) {
	dir := "../../../leetcode"
	number := resp.Data.Question.QuestionFrontendId
	d := filepath.Join(dir, number)
	if _, err := os.Stat(d); err != nil {
		err = os.Mkdir(d, os.ModePerm)
		if err != nil {
			panic(err)
			return
		}
	}
	solution := filepath.Join(dir, number, "solution.go")
	if _, err := os.Stat(solution); err != nil {
		create, err := os.Create(solution)
		if err != nil {
			panic(err)
			return
		}
		var code string
		for index := range resp.Data.Question.CodeSnippets {
			if resp.Data.Question.CodeSnippets[index].LangSlug == "golang" {
				code = resp.Data.Question.CodeSnippets[index].Code
			}
		}
		_, err = create.Write([]byte(fmt.Sprintf(solutionContent, code)))
		if err != nil {
			panic(err)
			return
		}
	}
	solution = filepath.Join(dir, number, "solution_test.go")
	if _, err := os.Stat(solution); err != nil {
		create, err := os.Create(solution)
		if err != nil {
			panic(err)
			return
		}
		_, err = create.Write([]byte(solutionTestContent))
		if err != nil {
			panic(err)
			return
		}
	}
	cmd := exec.Command("git", "add", d)
	cmd.Run()
}
