package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/gookit/color"
	"github.com/marguerite/util/command"
	"github.com/marguerite/util/dir"
	"github.com/marguerite/util/fileutils"
	"github.com/marguerite/util/slice"
)

var (
	reURL    = "(http(s)?:\\/\\/[^\\/]+\\/)?([^\\/]+\\/)?([^\\/@:]+)"
	reConf   = "(@([\\w]+)|\\/.*?\\/([^\\/]+))?\\/(.*?-packages.conf)"
	reRecipe = "(@([^:]+))?(:(\\w+))?(:(.*))?"
)

const (
	github = "https://github.com"
	rime   = "rime"
)

func eof() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}

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
		return recipes, errors.New("not a valid *-packages.conf")
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
	Name    string
	Options string
}

// URL print the Recipe's URL
func (r Recipe) URL() string {
	repo := r.Repo
	if r.Prefix == github && r.User == rime && !strings.HasPrefix(r.Repo, "rime-") {
		repo = "rime-" + repo
	}
	return r.Prefix+"/"+r.User+"/"+repo
}

// Recipes a slice of *Recipe
type Recipes []*Recipe

// NewRecipe generate a new Recipe object
func NewRecipe(prefix, user, repo, branch, name, options string) Recipe {
	if len(prefix) == 0 {
		prefix = github
	}
	if len(user) == 0 {
		user = rime
	}
	if len(branch) == 0 {
		branch = "master"
	}
	return Recipe{prefix, user, repo, branch, "", name, options}
}

func parseRecipes(strs []string) Recipes {
	recipes := Recipes{}
	re := regexp.MustCompile(reURL + reRecipe)
	for _, s := range strs {
		if !re.MatchString(s) {
			color.Warn.Printf("can't parse recipe URL %s"+eof(), s)
			continue
		}
		m := re.FindStringSubmatch(s)
		prefix := m[1]
		user := m[3]
		repo := m[4]
		branch := m[6]
		name := m[8]
		options := m[10]
		if len(prefix) == 0 {
			prefix = github
		}
		if len(user) == 0 {
			user = rime
		} else {
			user = strings.ReplaceAll(user, "/", "")
		}
		if len(branch) == 0 {
			branch = "master"
		}

		r := NewRecipe(prefix, user, repo, branch, name, options)
		recipes = append(recipes, &r)
	}
	return recipes
}

func remoteRecipeURL(conf string) (string, error) {
	re := regexp.MustCompile(reURL + reConf)
	if !re.MatchString(conf) {
		return conf, errors.New("not valid configuration URL")
	}
	m := re.FindStringSubmatch(conf)
	prefix := m[1]
	user := m[3]
	repo := m[4]
	path := m[5]
	pkgconf := m[8]
	if len(prefix) == 0 {
		prefix = github
	}
	if len(user) == 0 {
		user = rime
	} else {
		user = strings.ReplaceAll(user, "/", "")
	}
	if len(path) == 0 {
		path = "/raw/master/"
	}
	if strings.HasPrefix(path, "@") {
		path = "/raw/" + strings.ReplaceAll(path, "@", "") + "/"
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
		return recipes, fmt.Errorf("%s not fetched, status %d", url, resp.StatusCode)
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
	d, err := command.Environ(dir)
	if err == nil {
		return d
	}
	// guess based on operating system
	if runtime.GOOS == "windows" {
		d, _ := command.Environ("APPDATA")
		return filepath.Join(d, "Rime")
	}
	if runtime.GOOS == "darwin" {
		d, _ := command.Environ("HOME")
		return filepath.Join(d, "Library/Rime")
	}
	return getRimeDirLinux()
}

func getRimeDirLinux() string {
	im, err := command.Environ("GTK_IM_MODULE")
	if err != nil {
		// detect by binary
		for _, v := range []string{"ibus", "fcitx", "fcitx5"} {
			if val, err := command.Search(v); err == nil {
				return val
			}
		}
	}

	// system root?
	currentUser, _ := user.Current()
	if currentUser.Uid == "0" || currentUser.Username == "root" {
		return "/usr/share/rime-data"
	}

	home, _ := command.Environ("HOME")
	im = strings.Replace(im, "@im=", "", 1)
	switch im {
	case "fcitx":
		color.Info.Println("Installing for Rime Frontend: fcitx-rime")
		return filepath.Join(home, ".config/fcitx/rime")
	case "fcitx5":
		color.Info.Println("Installing for Rime Frontend: fcitx5-rime")
		return filepath.Join(home, ".config/fcitx5/rime")
	case "ibus":
		color.Info.Println("Installing for Rime Frontend: ibus-rime")
		return filepath.Join(home, ".config/ibus/rime")
	default:
		color.Warn.Printf("Unkown Rime Frontend: %s"+eof(), im)
		os.Exit(1)
	}
	return ""
}

func cloneOrUpdateRepos(urls Recipes, rimeDir string) {
	for _, v := range urls {
		pkgDir, _ := filepath.Abs(filepath.Join(rimeDir, "package", v.User, strings.ReplaceAll(v.Repo, "rime-", "")))
		v.Dir = pkgDir
		if _, err := os.Stat(pkgDir); os.IsNotExist(err) {
			color.Info.Printf("Fetching %s to %s"+eof(), v.Repo, pkgDir)
			_, err := git.PlainClone(pkgDir, false, &git.CloneOptions{
				URL:               v.URL(),
				ReferenceName:     plumbing.NewBranchReferenceName(v.Branch),
				RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
				Depth:             1})
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		} else {
			color.Info.Printf("Updating %s"+eof(), pkgDir)
			// Switch branch?
			/*r, err := git.PlainOpen(pkgDir)
			if err != nil {
				color.Error.Printf("%s is not a valid git repository"+eof(), pkgDir)
				os.Exit(1)
			}
			w, err := r.Worktree()
			if err != nil {
				color.Error.Printf("%s doesn't contains any worktree"+eof(), pkgDir)
				os.Exit(1)
			}
			err = w.Pull(&git.PullOptions{RemoteName: "origin"})
			if err != nil {
				if err.Error() == "already up-to-date" {
					color.Warn.Printf("%s already up-to-date"+eof(), v.URL())
				} else {
					color.Error.Printf("Failed to pull repository %s: %s"+eof(), v.URL(), err.Error())
					os.Exit(1)
				}
			}
			err = w.Checkout(&git.CheckoutOptions{
				Branch: plumbing.NewBranchReferenceName(v.Branch),
				Force:  true,
			})
			if err != nil {
				color.Error.Printf("Failed to checkout branch %s"+eof(), v.Branch)
				os.Exit(1)
			}*/
		}
	}
}

// RecipeConf contains useful information in the recipe file
type RecipeConf struct {
	Rx    string
	Args  map[string]string
	Files []string
	Patch string
}

func parseRecipeConf(recipeFile string, recipe Recipe) RecipeConf {
	if _, err := os.Stat(recipeFile); os.IsNotExist(err) {
		color.Error.Printf("Recipe not found %s"+eof(), recipeFile)
		os.Exit(1)
	}
	f, err := ioutil.ReadFile(recipeFile)
	if err != nil {
		color.Error.Printf("Can not read %s: %s"+eof(), recipeFile, err.Error())
		os.Exit(1)
	}
	re := regexp.MustCompile(`(?s)Rx:\s+(\w+).*?args:(.*?)description.*?(install_files:.*?\n(.*)\n\n)?patch_files:\n(.*)`)
	if !re.MatchString(string(f)) {
		color.Error.Printf("Recipe file's format is wrong: %s"+eof(), recipeFile)
		os.Exit(1)
	}
	m := re.FindStringSubmatch(string(f))
	if m[1] != recipe.Name {
		color.Error.Printf("Invalid Recipe %s doesn't match file name %s"+eof(), m[1], recipe.Name)
		os.Exit(1)
	}
	files := []string{}
	if len(m[3]) > 0 {
		files = strings.Split(strings.TrimSpace(m[4]), eof())
	}
	return RecipeConf{m[1], parseRecipeArgs(m[2], recipe.Options), files, strings.TrimSpace(m[5])}
}

func parseRecipeArgs(argStr string, optStr string) map[string]string {
	re := regexp.MustCompile(`(\w+)(=(.*))?`)
	args := map[string]string{}
	for _, v := range strings.Split(argStr, "-") {
		v = strings.TrimSpace(v)
		if len(v) > 0 && re.MatchString(v) {
			m := re.FindStringSubmatch(v)
			if len(m[2]) > 0 {
				args[m[1]] = m[3]
			} else {
				args[m[1]] = ""
			}
		}
	}
	opts := map[string]string{}
	for _, v := range strings.Split(optStr, ",") {
		if re.MatchString(v) {
			m := re.FindStringSubmatch(v)
			opts[m[1]] = m[3]
		}
	}

	for k := range args {
		if val, ok := opts[k]; ok {
			args[k] = val
		}
	}

	return args
}

func applyInstallFiles(files []string, src, dst string) {
	for _, f := range files {
		f = strings.TrimSpace(f)
		if strings.HasPrefix(f, "#") {
			continue
		}
		f = strings.TrimLeft(f, "- ")
		err := fileutils.Copy(filepath.Join(src, f), dst)
		if err != nil {
			color.Error.Printf("Failed to copy %s to %s"+eof(), filepath.Join(src, f), dst)
			os.Exit(1)
		}
	}
}

func replacePatchVariable(s string, m [][]string, args map[string]string) string {
	for _, v := range m {
		str := ""
		var re *regexp.Regexp
		switch v[3] {
		case "%":
			re = regexp.MustCompile("(.*?)" + strings.ReplaceAll(v[4], "\\", "\\\\") + ".*?$")
		case "%%":
			re = regexp.MustCompile("(.*?)" + strings.ReplaceAll(v[4], "\\", "\\\\") + ".*$")
		case "#":
			re = regexp.MustCompile("^.*?" + strings.ReplaceAll(v[4], "\\", "\\\\") + "(.*?)")
		case "##":
			re = regexp.MustCompile("^.*" + strings.ReplaceAll(v[4], "\\", "\\\\") + "(.*?)")
		default:
			re = regexp.MustCompile("")
		}

		if len(re.String()) > 0 && re.MatchString(args[v[1]]) {
			str = re.FindStringSubmatch(args[v[1]])[1]
		}
		if v[3] == ":-" && len(args[v[1]]) == 0 {
			str = v[4]
		}
		if len(str) > 0 {
			s = strings.Replace(s, v[0], str, 1)
		} else {
			s = strings.Replace(s, v[0], args[v[1]], 1)
		}
	}
	return s
}

// replace args variable to string in patchContent
func parsePatch(content string, args map[string]string, recipe Recipe) map[string]string {
	// split content
	rd := bufio.NewScanner(strings.NewReader(content))
	contents := []string{}
	tmp := ""
	idx := 0
	for rd.Scan() {
		if strings.HasSuffix(rd.Text(), ".yaml:") {
			idx++
		}
		if idx > 1 {
			contents = append(contents, tmp)
			tmp = ""
			idx = 1
		}
		tmp += rd.Text() + eof()
	}
	// write the last part
	contents = append(contents, tmp)

	patches := map[string]string{}

	for _, c := range contents {
		reader := bufio.NewScanner(strings.NewReader(c))
		patch := "# RX: " + recipe.Repo + ":" + recipe.Name + ":" + recipe.Options + " {\n"
		file := ""
		re := regexp.MustCompile(`\$\{(.*?)(([%#:-]+)(.*?))?\}`)
		i := 0
		sep := "\t"
		for reader.Scan() {
			line := strings.TrimSpace(reader.Text())
			str := ""
			if re.MatchString(line) {
				m := re.FindAllStringSubmatch(line, -1)
				str += replacePatchVariable(line, m, args)
			} else {
				str += line
			}
			if i == 0 {
				file += strings.Replace(str, ":", "", 1)
			} else {
				patch += sep + str + eof()
				sep = strings.Repeat(sep, 2)
			}
			i++
		}
		patches[file] = patch + "# }" + eof()
	}
	return patches
}

func installPackages(recipes Recipes, dst string) {
	for _, recipe := range recipes {
		if len(recipe.Name) != 0 {
			installRecipe(filepath.Join(recipe.Dir, recipe.Name+".recipe.yaml"), dst, *recipe)
			continue
		}
		if _, err := os.Stat(filepath.Join(recipe.Dir, "recipe.yaml")); !os.IsNotExist(err) {
			installRecipe(filepath.Join(recipe.Dir, "recipe.yaml"), dst, *recipe)
			continue
		}
		installFilesFromDir(recipe.Dir, dst)
	}
}

func installRecipe(recipeFile, dst string, recipe Recipe) {
	rx := recipe.Repo + "/" + recipe.Name
	color.Info.Printf("Installing recipe: %s"+eof(), rx)
	color.Info.Printf("\t- option: %s"+eof(), recipe.Options)
	r := parseRecipeConf(recipeFile, recipe)
	applyInstallFiles(r.Files, filepath.Dir(recipeFile), dst)
	patches := parsePatch(r.Patch, r.Args, recipe)
	fmt.Println(patches)
	for k, v := range patches {
		dst = filepath.Join(dst, k)
		fmt.Println(dst)
		if _, err := os.Stat(dst); os.IsNotExist(err) {
			err = ioutil.WriteFile(dst, []byte("__patch:\n"+v), 0644)
			if err != nil {
				color.Error.Printf("Failed to write %s"+eof(), dst)
				os.Exit(1)
			}
		} else {
			f, err := ioutil.ReadFile(dst)
			if err != nil {
				color.Error.Printf("Failed to read existing %s"+eof(), dst)
				os.Exit(1)
			}
			re := regexp.MustCompile("(?s)# RX:.*?{\n.*}\n")
			if re.MatchString(string(f)) {
				err := ioutil.WriteFile(dst, []byte(strings.Replace(string(f), re.FindStringSubmatch(string(f))[0], v, 1)), 0644)
				if err != nil {
					color.Error.Printf("Failed to write %s"+eof(), dst)
					os.Exit(1)
				}
			} else {
				color.Error.Printf("%s doesn't contain any recipe/patch"+eof(), dst)
				os.Exit(1)
			}
		}
	}
}

func installFilesFromDir(d, dst string) {
	pattern := []string{".*\\.yaml", ".*\\.txt", ".*\\.gram", "opencc\\/.*\\..*"}
	ex := []string{"(custom|recipe)\\.yaml", "opencc\\/", "", "\\.(json|ocd|txt)"}
	files, err := dir.Glob(d, pattern, ex)
	if err != nil {
		color.Error.Printf("Can not find qualified files in %s"+eof(), d)
		os.Exit(1)
	}
	for _, v := range files {
		err := fileutils.Copy(v, dst)
		if err != nil {
			color.Error.Printf("Failed to copy %s to %s"+eof(), v, dst)
			os.Exit(1)
		}
	}
}

func buildPackages(d string) {
	files, _ := dir.Ls(d)
	if ok, _ := slice.Contains(files, "essay.txt"); ok {
		minEssay(filepath.Join(d, "essay.txt"))
	}
	if ok, _ := slice.Contains(files, "luna_pinyin.dict.yaml"); ok {
		minLuna(filepath.Join(d, "luna_pinyin.dict.yaml"))
	}
	schemas, _ := dir.Glob(d, regexp.MustCompile(`.*\.schema.yaml$`))
	for _, v := range schemas {
		minSchema(v)
	}
	overrideDefaultYaml(schemas)
	_, _, err := command.Run("/usr/bin/rime_deployer", "--build", d)
	if err != nil {
		color.Error.Printf("Failed to run command: %s"+eof(), "/usr/bin/rime_deployer --build "+d)
	}
	err = os.RemoveAll(filepath.Join(d, "user.yaml"))
	if err != nil {
		color.Error.Printf("Failed to remove %s"+eof(), filepath.Join(d, "user.yaml"))
	}
}

func overrideDefaultYaml(schemas []string) {
	m := []string{}
	for _, v := range schemas {
		s := strings.TrimRight(filepath.Base(v), ".schema.yaml")
		m = append(m, "- schema: "+s)
	}

	d := filepath.Join(filepath.Dir(schemas[0]), "default.yaml")
	f, _ := ioutil.ReadFile(d)
	re := regexp.MustCompile(`^config_version:\s+\'(.*)\'$`)
	scanner := bufio.NewScanner(strings.NewReader(string(f)))
	list := ""
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "- schema:") {
			for _, i := range m {
				if strings.Contains(line, i) {
					list += line + eof()
					continue
				}
			}
		} else {
			if re.MatchString(line) {
				list += "config_version: '" + re.FindStringSubmatch(line)[1] + ".minimal'" + eof()
			} else {
				list += line + eof()
			}
		}
	}

	err := ioutil.WriteFile(d, []byte(list), 0644)
	if err != nil {
		color.Error.Printf("Failed to write %s"+eof(), d)
	}
}

func minEssay(s string) {
	f, _ := ioutil.ReadFile(s)
	essay := ""
	scanner := bufio.NewScanner(strings.NewReader(string(f)))
	re := regexp.MustCompile(`^.*\s+(.*)$`)
	for scanner.Scan() {
		line := scanner.Text()
		if re.MatchString(line) {
			idx := re.FindStringSubmatch(line)[1]
			i, _ := strconv.Atoi(idx)
			if i >= 500 {
				essay += line + eof()
			}
		}
	}
	err := ioutil.WriteFile(s, []byte(essay), 0644)
	if err != nil {
		color.Error.Printf("Failed to write %s"+eof(), s)
	}
}

func minLuna(s string) {
	f, _ := ioutil.ReadFile(s)
	scanner := bufio.NewScanner(strings.NewReader(string(f)))
	luna := ""
	re := regexp.MustCompile(`^version:\s+\"(.*)\"$`)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#以下") {
			break
		}
		if re.MatchString(line) {
			luna += "version:\t\"" + re.FindStringSubmatch(line)[1] + ".minimal\"" + eof()
		} else {
			luna += line + eof()
		}
	}
	err := ioutil.WriteFile(s, []byte(luna), 0644)
	if err != nil {
		color.Error.Printf("Failed to write file %s"+eof(), s)
	}
}

func minSchema(s string) {
	f, _ := ioutil.ReadFile(s)
	scanner := bufio.NewScanner(strings.NewReader(string(f)))
	schema := ""
	re := regexp.MustCompile(`(\s+version:\s+)\"(.*)\"$`)
	for scanner.Scan() {
		line := scanner.Text()
		if re.MatchString(line) {
			m := re.FindStringSubmatch(line)
			schema += m[1] + "\"" + m[2] + ".minimal\"" + eof()
			continue
		}
		if strings.Contains(line, "- stroke") {
			schema += "#" + line + eof()
			continue
		}
		if strings.Contains(line, "- reverse_lookup_translator") {
			schema += "#" + line + eof()
		}
		schema += line + eof()
	}
	err := ioutil.WriteFile(s, []byte(schema), 0644)
	if err != nil {
		color.Error.Printf("Failed to write file %s"+eof(), s)
	}
}

func main() {
	var recipeStr, rimeDir string
	var build bool
	flag.StringVar(&recipeStr, "r", "", "pass recipe url and commands.")
	flag.StringVar(&rimeDir, "d", "", "where to install recipes.")
	flag.BoolVar(&build, "b", false, "whether to build packages.")
	flag.Parse()

	if len(rimeDir) == 0 {
		rimeDir = getDir("RIME_DIR")
	}

	recipeStrs := parseRecipeStrs(strings.Split(recipeStr, " "))
	recipes := parseRecipes(recipeStrs)
	cloneOrUpdateRepos(recipes, rimeDir)
	installPackages(recipes, rimeDir)
	if build {
		buildPackages(rimeDir)
	}
}
