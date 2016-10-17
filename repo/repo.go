package repo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"

	"github.com/G-Node/gin-cli/auth"
	"github.com/G-Node/gin-cli/client"
	"github.com/G-Node/gin-repo/wire"
)

const repohost = "https://repo.gin.g-node.org"

// GetRepos gets a list of repositories (public or user specific)
func GetRepos(user, token string) ([]wire.Repo, error) {
	repocl := client.NewClient(repohost)
	var res *http.Response
	var err error
	if user == "" {
		res, err = repocl.Get("/repos/public")
	} else {
		repocl.Token = token
		res, err = repocl.Get(fmt.Sprintf("/users/%s/repos", user))
	}

	if err != nil {
		return nil, fmt.Errorf("Request for repositories returned error: %s", err)
	} else if res.StatusCode != 200 {
		return nil, fmt.Errorf("[Repository request error] Server returned: %s", res.Status)
	}

	defer client.CloseRes(res.Body)
	b, err := ioutil.ReadAll(res.Body)
	var repoList []wire.Repo
	err = json.Unmarshal(b, &repoList)
	if err != nil {
		return nil, err
	}

	return repoList, nil
}

// CreateRepo creates a repository on the server.
func CreateRepo(name, description string) error {
	repocl := client.NewClient(repohost)
	username, token, _ := auth.LoadToken(false)
	repocl.Token = token

	data := wire.Repo{Name: name, Description: description}
	res, err := repocl.Post(fmt.Sprintf("/users/%s/repos", username), data)
	if err != nil {
		return err
	}
	defer client.CloseRes(res.Body)
	if res.StatusCode != 201 {
		return fmt.Errorf("Failed to create repository. Server returned: %s", res.Status)
	}

	return nil
}

// UploadRepo adds files to a repository and uploads them.
func UploadRepo(localPath string) error {
	defer CleanUpTemp()

	idx, err := AddPath(localPath)
	if err != nil {
		return err
	}

	err = Commit(localPath, idx)
	if err != nil {
		return err
	}

	err = Push(localPath)
	if err != nil {
		return err
	}

	return AnnexPush(localPath)
}

// DownloadRepo downloads the files in an already checked out repository.
func DownloadRepo() error {
	defer CleanUpTemp()
	// git pull
	err := Pull()
	if err != nil {
		return err
	}

	// git annex pull
	return AnnexPull(".")

}

// CloneRepo downloads the files of a given repository.
func CloneRepo(repoPath string) error {
	defer CleanUpTemp()

	localPath := path.Base(repoPath)
	fmt.Printf("Fetching repository '%s'... ", localPath)
	_, err := Clone(repoPath)
	if err != nil {
		fmt.Printf("failed!\n")
		return err
	}
	fmt.Printf("done.\n")

	fmt.Printf("Downloading files... ")
	err = AnnexPull(localPath)
	if err != nil {
		fmt.Printf("failed!\n")
		return err
	}
	fmt.Printf("done.\n")
	return nil
}