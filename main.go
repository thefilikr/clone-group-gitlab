package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/ilyakaznacheev/cleanenv"
)

type Settings struct {
	GitLabURL    string `yaml:"gitlab_url"` // GitLab URL
	GroupID      string `yaml:"group_id"`   // Group ID
	PrivateToken string `yaml:"token"`      // Personal access token
	CloneDir     string `yaml:"clone_dir"`  // Directory for cloning
	PerPage      int    `yaml:"per_page"`   // Number of items per page
}

func MustLoad() *Settings {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Settings

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("error reading config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = "./config.yaml"
	}

	return res
}

type Project struct {
	HTTPURLToRepo     string `json:"http_url_to_repo"`
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
}

type Subgroup struct {
	ID   int    `json:"id"`
	Path string `json:"path"`
}

func getProjects(cfg *Settings, groupID string) ([]Project, error) {
	var projects []Project
	page := 1

	for {
		url := fmt.Sprintf("%s/api/v4/groups/%s/projects?per_page=%d&page=%d", cfg.GitLabURL, groupID, cfg.PerPage, page)
		resp, err := sendRequest(cfg, url)
		if err != nil {
			return nil, err
		}

		var batch []Project
		if err := json.Unmarshal(resp, &batch); err != nil {
			return nil, err
		}
		if len(batch) == 0 {
			break
		}
		projects = append(projects, batch...)
		page++
	}

	return projects, nil
}

func getSubgroups(cfg *Settings, groupID string) ([]Subgroup, error) {
	url := fmt.Sprintf("%s/api/v4/groups/%s/subgroups", cfg.GitLabURL, groupID)
	resp, err := sendRequest(cfg, url)
	if err != nil {
		return nil, err
	}

	var subgroups []Subgroup
	if err := json.Unmarshal(resp, &subgroups); err != nil {
		return nil, err
	}

	return subgroups, nil
}

func getAllProjects(cfg *Settings, groupID string) ([]Project, error) {
	projects, err := getProjects(cfg, groupID)
	if err != nil {
		return nil, err
	}

	subgroups, err := getSubgroups(cfg, groupID)
	if err != nil {
		return nil, err
	}

	for _, subgroup := range subgroups {
		subProjects, err := getAllProjects(cfg, strconv.Itoa(subgroup.ID))
		if err != nil {
			return nil, err
		}
		projects = append(projects, subProjects...)
	}

	return projects, nil
}

func sendRequest(cfg *Settings, url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", cfg.PrivateToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func cloneProjects(cfg *Settings, projects []Project) error {
	for _, project := range projects {
		projectPath := filepath.Join(cfg.CloneDir, project.PathWithNamespace)

		if err := os.MkdirAll(filepath.Dir(projectPath), 0755); err != nil {
			return err
		}

		repoURL := addTokenToURL(project.HTTPURLToRepo, cfg.PrivateToken)
		fmt.Printf("Cloning %s into %s...\n", project.Name, projectPath)

		cmd := exec.Command("git", "clone", repoURL, projectPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("Failed to clone %s: %v\n", project.Name, err)
		}

		fmt.Println()
	}

	return nil
}

func addTokenToURL(repoURL, token string) string {
	return fmt.Sprintf("https://%s@%s", token, repoURL[8:])
}

func main() {
	cfg := MustLoad()

	fmt.Println("Fetching all projects...")
	projects, err := getAllProjects(cfg, cfg.GroupID)
	if err != nil {
		fmt.Printf("Error fetching projects: %v\n", err)
		return
	}

	fmt.Printf("Found %d projects.\n", len(projects))

	if err := cloneProjects(cfg, projects); err != nil {
		fmt.Printf("Error cloning projects: %v\n", err)
		return
	}

	fmt.Println("All projects cloned successfully.")
}
