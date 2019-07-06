package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/marguerite/util/dirutils"
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

// Recipe an object contains everything github clone needs
type Recipe struct {
	Prefix  string
	User    string
	Repo    string
	Branch  string
	Dir     string
	Action  string
	Options string
}

// URL print the Recipe's URL
func (r Recipe) URL() string {
	repo := r.Repo
	if r.Prefix == "https://github.com/" && r.User == "rime" && !strings.HasPrefix(r.Repo, "rime-") {
		repo = "rime-" + repo
	}
	return r.Prefix + r.User + "/" + repo
}

func (r *Recipe) SetDir(d string) {
	r.Dir = d
}

// Recipes a slice of *Recipe
type Recipes []*Recipe

// NewRecipe generate a new Recipe object
func NewRecipe(prefix, user, repo, branch, action, options string) Recipe {
	if len(prefix) == 0 {
		prefix = "https://github.com/"
	}
	if len(user) == 0 {
		user = "rime"
	}
	if len(branch) == 0 {
		branch = "master"
	}
	return Recipe{prefix, user, repo, branch, "", action, options}
}

func parseRecipes(strs []string) Recipes {
	recipes := Recipes{}
	re := regexp.MustCompile(reURL + reRecipe)
	for _, s := range strs {
		if !re.MatchString(s) {
			fmt.Printf("can't parse recipe URL %s.\n", s)
			continue
		}
		m := re.FindStringSubmatch(s)
		prefix := m[1]
		user := m[3]
		repo := m[4]
		branch := m[6]
		action := ""
		options := ""
		if len(prefix) == 0 {
			prefix = "https://github.com/"
		}
		if len(user) == 0 {
			user = "rime"
		} else {
			user = strings.Replace(user, "/", "", -1)
		}
		if len(branch) == 0 {
			branch = "master"
		}
		if len(m[7]) != 0 {
			action = m[8]
			options = m[9]
		}

		r := NewRecipe(prefix, user, repo, branch, action, options)
		recipes = append(recipes, &r)
	}
	return recipes
}

func remoteRecipeURL(conf string) (string, error) {
	re := regexp.MustCompile(reURL + reConf)
	if !re.MatchString(conf) {
		return conf, errors.New("Not valid configuration URL.")
	}
	m := re.FindStringSubmatch(conf)
	prefix := m[1]
	user := m[3]
	repo := m[4]
	path := m[5]
	pkgconf := m[8]
	if len(prefix) == 0 {
		prefix = "https://github.com/"
	}
	if len(user) == 0 {
		user = "rime"
	} else {
		user = strings.Replace(user, "/", "", -1)
	}
	if len(path) == 0 {
		path = "/raw/master/"
	}
	if strings.HasPrefix(path, "@") {
		path = "/raw/" + strings.Replace(path, "@", "", -1) + "/"
	}
	url := prefix + user + "/" + repo + path + pkgconf
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

func parseRecipeStrs(args []string) []string {
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

func cloneOrUpdateRepos(urls Recipes, rimeDir string) {
	if len(rimeDir) == 0 {
		rimeDir = getDir("RIME_DIR")
	}
	for _, v := range urls {
		pkgDir, _ := filepath.Abs(rimeDir + "/package/" + v.User + "/" + strings.Replace(v.Repo, "rime-", "", -1))
		v.SetDir(pkgDir)
		if _, err := os.Stat(pkgDir); os.IsNotExist(err) {
			fmt.Printf("Cloneing %s to %s.\n", v.Repo, pkgDir)
			_, err := git.PlainClone(pkgDir, false, &git.CloneOptions{
				URL:           v.URL(),
				ReferenceName: plumbing.NewBranchReferenceName(v.Branch),
				SingleBranch:  true})
			fmt.Println(err)
		} else {
			fmt.Printf("Updating %s.\n", pkgDir)
			r, err := git.PlainOpen(pkgDir)
			if err != nil {
				// not git working dir
			}
			w, err := r.Worktree()
			if err != nil {
			}
			err = w.Pull(&git.PullOptions{RemoteName: "origin"})
			if err != nil {
				if err.Error() == "already up-to-date" {
					fmt.Printf("%s already up-to-date.\n", v.URL())
					continue
				}
				fmt.Printf("Failed to pull repository %s: %s.\n", v.URL(), err.Error())
				os.Exit(1)
			}
		}
	}
}

func installPackages(recipes Recipes) {
	for _, recipe := range recipes {
		if len(recipe.Options) != 0 {
			installRecipe(filepath.Join(recipe.Dir, recipe.Repo+".recipe.yaml"))
			continue
		}
		if _, err := os.Stat(filepath.Join(recipe.Dir, "recipe.yaml")); !os.IsNotExist(err) {
			installRecipe(filepath.Join(recipe.Dir, "recipe.yaml"))
			continue
		}
		installFilesFromDir(recipe.Dir)
	}
}

func installRecipe(recipe string) {
	fmt.Println(recipe)
}

func installFilesFromDir(dir string) {
	pattern := []string{".*\\.yaml", ".*\\.txt", ".*\\.gram", "opencc\\/.*\\..*"}
	ex := []string{"(custom|recipe)\\.yaml", "opencc\\/", "", "\\.(json|ocd|txt)"}
	files, err := dirutils.Glob(dir, pattern, ex)
	fmt.Println(files)
	fmt.Println(err)
}

func main() {
	var recipeStr, rimeDir string
	flag.StringVar(&recipeStr, "r", "", "pass recipe url and commands.")
	flag.StringVar(&rimeDir, "d", "", "where to install recipes.")
	flag.Parse()

	recipeStrs := parseRecipeStrs(strings.Split(recipeStr, " "))
	recipes := parseRecipes(recipeStrs)
	cloneOrUpdateRepos(recipes, rimeDir)
	installPackages(recipes)
}