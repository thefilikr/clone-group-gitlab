
# GitLab Group Cloner

This Go script allows you to automatically clone all repositories from a GitLab group and its subgroups into the local file system, preserving the original subgroup hierarchy as directories.

## Features

- **Cloning via HTTPS**: The script uses HTTPS to clone repositories.
- **Preserving structure**: The subgroup hierarchy from GitLab is converted into corresponding directory structure.
- **Automatic scanning**: The script recursively collects information about all repositories in the group and its subgroups.

## Example Structure

### Original structure in GitLab

```
Group
├── Subgroup1
│   ├── Repo1
│   └── Repo2
└── Subgroup2
    └── Repo3
```

### Local structure after cloning

```
cloneDir/
├── Group/
│   ├── Subgroup1/
│   │   ├── Repo1/
│   │   └── Repo2/
│   └── Subgroup2/
│       └── Repo3/
```

## Installation

1. Make sure you have Go version 1.16 or higher installed.
2. Clone this repository:
   ```bash
   git clone https://your-repository-url.git
   ```
3. Install dependencies:
   ```bash
   go mod tidy
   ```

## Configuration

Create a `config.yaml` file in the root directory with the following content:

```yaml
gitlab_url: "https://gitlab.example.com"
group_id: "123456" # Group ID
token: "your-private-token" # Personal access token
clone_dir: "./repositories" # Directory for cloning
per_page: 50 # Number of items per page
```

## Running

Run the following command to start the script:

```bash
go run main.go --config=config.yaml
```

## Example Result

### Screenshots

On the left, you can see the structure of groups and subgroups in GitLab, and on the right, the corresponding local directory structure after cloning.

<p align="center">
  <img src="./images/gitlab_structure.png" alt="GitLab Structure" width="45%">
  <img src="./images/local_structure.png" alt="Local Structure" width="45%">
</p>

## Notes

1. Make sure the access token has sufficient permissions to read repositories and subgroups.
2. Git must be installed for the script to work.

