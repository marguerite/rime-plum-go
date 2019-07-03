package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/marguerite/util/slice"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	reURL    string = "(http(s)?:\\/\\/[^\\/]+\\/)?([^\\/]+\\/)?([^\\/@:]+)"
	reConf   string = "(@([\\w]+)|\\/.*?\\/([^\\/]+))?\\/(.*?-packages.conf)"
	reRecipe string = "(@([^:]+))?(:(\\w+):(.*))?"
)

func presetRecipes() []string {
	return []string{"bopomofo", "cangjie", "essay", "luna-pinyin", "prelude", "stroke", "terra-pinyin"}
}

func extraRecipes() []string {
	return []string{"array", "combo-pinyin", "double-pinyin", "emoji", "ipa", "jyutping", "middle-chinese", "pinyin-simp", "quick", "scj", "soutzoe", "stenotype", "wubi", "wugniu"}
}

func allRecipes() []string {
	r := presetRecipes()
	slice.Concat(&r, extraRecipes())
	return r
}

func parsePackagesConf(conf string) ([]string, error) {
	recipes := []string{}
	re := regexp.MustCompile(`(?s)package_list(\+)?=\(\n(.*?)\)`)
	if !re.MatchString(conf) {
		return recipes, errors.New("Not a valid *-packages.conf.")
	}
	scanner := bufio.NewScanner(strings.NewReader(re.FindStringSubmatch(conf)[2]))
	for scanner.Scan() {
		recipes = append(recipes, strings.TrimSpace(scanner.Text()))
	}
	return recipes, nil
}

func genPredefinedRecipes(r string, recipes []string) []string {
	if r == ":preset" {
		slice.Concat(&recipes, presetRecipes())
	}
	if r == ":extra" {
		slice.Concat(&recipes, extraRecipes())
	}
	if r == ":all" {
		slice.Concat(&recipes, allRecipes())
	}
	return recipes
}

// RecipeURL an object contains everything github clone needs
type RecipeURL struct {
	Prefix string
	Author string
	Repo   string
	Branch string
}

// String print the RecipeURL as string
func (r RecipeURL) String() string {
	return r.Prefix + r.Author + "/" + r.Repo
}

// RecipeURLs a slice of RecipeURL
type RecipeURLs []RecipeURL

// NewRecipeURL generate a new RecipeURL object
func NewRecipeURL(prefix, author, repo, branch string) RecipeURL {
	if len(prefix) == 0 {
		prefix = "https://github.com/"
	}
	if len(author) == 0 {
		author = "rime"
	}
	if len(branch) == 0 {
		branch = "master"
	}
	if prefix == "https://github.com" && author == "rime" && strings.HasPrefix(repo, "rime-") {
		repo = "rime-" + repo
	}
	return RecipeURL{prefix, author, repo, branch}
}

func parseRecipeUrl(recipes []string) RecipeURLs {
	urls := RecipeURLs{}
	re := regexp.MustCompile(reURL + reRecipe)
	for _, recipe := range recipes {
		if !re.MatchString(recipe) {
			fmt.Printf("can't parse recipe URL %s.\n", recipe)
			continue
		}
		m := re.FindStringSubmatch(recipe)
		prefix := m[1]
		author := m[3]
		repo := m[4]
		branch := m[6]
		if len(prefix) == 0 {
			prefix = "https://github.com/"
		}
		if len(author) == 0 {
			author = "rime"
		} else {
			author = strings.Replace(author, "/", "", -1)
		}
		if len(branch) == 0 {
			branch = "master"
		}
		urls = append(urls, NewRecipeURL(prefix, author, repo, branch))
	}
	return urls
}

func remoteRecipeURL(conf string) (string, error) {
	re := regexp.MustCompile(reURL + reConf)
	if !re.MatchString(conf) {
		return conf, errors.New("Not valid configuration URL.")
	}
	m := re.FindStringSubmatch(conf)
	prefix := m[1]
	author := m[3]
	repo := m[4]
	path := m[5]
	pkgconf := m[8]
	if len(prefix) == 0 {
		prefix = "https://github.com/"
	}
	if len(author) == 0 {
		author = "rime"
	} else {
		author = strings.Replace(author, "/", "", -1)
	}
	if len(path) == 0 {
		path = "/raw/master/"
	}
	if strings.HasPrefix(path, "@") {
		path = "/raw/" + strings.Replace(path, "@", "", -1) + "/"
	}
	url := prefix + author + "/" + repo + path + pkgconf
	return url, nil
}

func fetchRemoteRecipes(conf string) ([]string, error) {
	recipes := []string{}
	url, err := remoteRecipeURL(conf)
	if err != nil {
		return recipes, err
	}

	resp, err := http.Get(url)
	if err != nil {
		return recipes, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return recipes, fmt.Errorf("%s not fetched, status %d.", url, resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return recipes, err
	}

	o, err := parsePackagesConf(string(b))
	if err != nil {
		return recipes, err
	}

	return o, nil
}

func getEnv(env string) (string, error) {
	val, ok := os.LookupEnv(env)
	if !ok {
		return "", fmt.Errorf("%s not set.", env)
	}
	if len(val) == 0 {
		return val, fmt.Errorf("%s is empty.", env)
	}
	return val, nil
}

func parseRecipes(args []string) []string {
	recipes := []string{}
	for _, recipe := range args {
		// predefined recipes
		if strings.HasPrefix(recipe, ":") {
			slice.Concat(&recipes, genPredefinedRecipes(recipe, recipes))
			continue
		}
		if strings.HasSuffix(recipe, ".conf") {
			r, err := fetchRemoteRecipes(recipe)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			slice.Concat(&recipes, r)
			continue
		}
		recipes = append(recipes, recipe)
	}
	return recipes
}

func getDir(dir string) string {
	d, err := getEnv(dir)
	if err == nil {
		return d
	}
	// guess based on operating system
	return ""
}

func cloneOrUpdateRepos(urls RecipeURLs) {
	rimeDir := getDir("RIME_DIR")
	for _, v := range urls {
		_, err := git.PlainClone(filepath.Join(rimeDir, v.Repo), false, &git.CloneOptions{
			URL:           v.String(),
			ReferenceName: plumbing.NewBranchReferenceName(v.Branch),
			SingleBranch:  true})
		fmt.Println(err)
	}
}

func main() {
	var recipeStr string
	flag.StringVar(&recipeStr, "r", "", "pass recipe url and commands.")
	flag.Parse()

	recipes := parseRecipes(strings.Split(recipeStr, " "))
	urls := parseRecipeUrl(recipes)
	cloneOrUpdateRepos(urls)
}
