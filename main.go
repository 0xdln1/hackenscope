package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36"

type ProgramListResponse struct {
	Programs []struct {
		Slug string `json:"slug"`
	} `json:"programs"`
}

type ProgramDetail struct {
	Slug   string `json:"slug"`
	Scopes []struct {
		Title       string `json:"title"`
		Target      string `json:"target"`
		OutOfScope  bool   `json:"out_of_scope"`
		RewardType  string `json:"reward_type"`
		Criticality string `json:"criticality"`
		TargetDesc  string `json:"target_description"`
	} `json:"scopes"`
}

func doRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	return http.DefaultClient.Do(req)
}

func fetchProgramSlugs(page int) ([]string, error) {
	url := fmt.Sprintf("https://hackenproof.com/programs-api/programs?not_audits=true&search&page=%d&order_by[published_date]=desc&with_abilities[]=Web", page)
	resp, err := doRequest(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var list ProgramListResponse
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, err
	}

	var slugs []string
	for _, p := range list.Programs {
		slugs = append(slugs, p.Slug)
	}
	return slugs, nil
}

func fetchProgramDetails(slug string) (*ProgramDetail, error) {
	url := fmt.Sprintf("https://hackenproof.com/programs-api/programs/%s", slug)
	resp, err := doRequest(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var detail ProgramDetail
	if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
		return nil, err
	}
	return &detail, nil
}

func main() {
	page := 1
	for {
		slugs, err := fetchProgramSlugs(page)
		if err != nil {
			log.Printf("Error fetching program slugs on page %d: %v", page, err)
			break
		}
		if len(slugs) == 0 {
			break
		}
		for _, slug := range slugs {
			detail, err := fetchProgramDetails(slug)
			if err != nil {
				log.Printf("Error fetching details for slug %s: %v", slug, err)
				continue
			}
			for _, scope := range detail.Scopes {
				if !scope.OutOfScope {
					fmt.Printf(
						"%s, https://hackenproof.com/programs/%s, %s, %s, %s, %s\n",
						scope.Target,
						detail.Slug,
						scope.Title,
						scope.Criticality,
						scope.RewardType,
						scope.TargetDesc,
					)
				}
			}
			time.Sleep(500 * time.Millisecond)
		}
		page++
	}
}
