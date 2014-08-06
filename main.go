package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"code.google.com/p/go-html-transform/css/selector"
	"code.google.com/p/go-html-transform/h5"
)

var backgroundColorRegexp = regexp.MustCompile(`background-color:(.*?)[";]`)

func get(u string) (resp *http.Response, err error) {
	log.Printf("GET %s", u)
	time.Sleep(time.Second * 8)
	return http.Get(u)
}

func fetchLanguages() ([]string, error) {
	resp, err := get("https://github.com/trending")
	if err != nil {
		return nil, err
	}
	tree, err := h5.New(resp.Body)
	if err != nil {
		return nil, err
	}
	chain, err := selector.Selector(".select-menu-list [data-filterable-for] .select-menu-item a")
	if err != nil {
		return nil, err
	}
	languages := []string{}
	for _, n := range chain.Find(tree.Top()) {
		if n.FirstChild != nil {
			languages = append(languages, n.FirstChild.Data)
		}
	}
	return languages, nil
}

func findProjectWithLanguage(language string) (string, error) {
	u := fmt.Sprintf("https://github.com/search?q=%s",
		url.QueryEscape(fmt.Sprintf(`language:"%s"`, language)))
	resp, err := get(u)
	if err != nil {
		return "", err
	}
	tree, err := h5.New(resp.Body)
	if err != nil {
		return "", err
	}
	chain, err := selector.Selector(".repolist-name a")
	if err != nil {
		return "", err
	}
	for _, n := range chain.Find(tree.Top()) {
		if n.FirstChild != nil {
			return n.FirstChild.Data, nil
		}
	}
	return "", fmt.Errorf("could not find project for %s", language)
}

func getProjectLanguageColors(project string) (map[string]string, error) {
	u := fmt.Sprintf("https://github.com/%s", project)
	resp, err := get(u)
	if err != nil {
		return nil, fmt.Errorf("could not get project page, %v", err)
	}
	tree, err := h5.New(resp.Body)
	if err != nil {
		return nil, err
	}
	chain, err := selector.Selector("span.language-color")
	if err != nil {
		return nil, err
	}
	colors := map[string]string{}
	for _, n := range chain.Find(tree.Top()) {
		if n.FirstChild != nil {
			for _, a := range n.Attr {
				if a.Key == "style" {
					matches := backgroundColorRegexp.FindStringSubmatch(a.Val)
					if matches != nil {
						colors[n.FirstChild.Data] = matches[1]
						log.Printf("Found color for %s: %s",
							n.FirstChild.Data, matches[1])
						break
					}
				}
			}
		}
	}
	return colors, nil

	return nil, errors.New("not implemented")
}

func parseHexColor(hex string) (color.RGBA, error) {
	return color.RGBA{}, errors.New("not implemented")
}

func main() {
	colors := map[string]string{}
	languages, err := fetchLanguages()
	if err != nil {
		log.Fatalf("Could not fetch languages, %v", err)
	}
	for _, lang := range languages {
		if colors[lang] != "" {
			log.Printf("Color already found for %s, skipping", lang)
			continue
		}
		log.Printf("Finding project for %s", lang)
		project, err := findProjectWithLanguage(lang)
		if err != nil {
			log.Fatalf("Could not find project for %s, %v", lang, err)
		}
		log.Printf("Found project for %s: %s", lang, project)
		projColors, err := getProjectLanguageColors(project)
		if err != nil {
			log.Fatalf("Could not find colors for project %s, %v", project, err)
		}
		for l, c := range projColors {
			if c != "" || colors[l] == "" {
				colors[l] = c
			}
		}
	}
	j, err := json.MarshalIndent(colors, "", "	")
	if err != nil {
		log.Fatalf("Could not convert to JSON, %v", err)
	}
	fmt.Println(string(j))
}
