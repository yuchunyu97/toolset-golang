package acr

// BuildRule BuildRule
type BuildRule struct {
	ImageTag           string `json:"imageTag"`
	DockerfileLocation string `json:"dockerfileLocation"`
	DockerfileName     string `json:"dockerfileName"`
	PushType           string `json:"pushType"`
	PushName           string `json:"pushName"`
	BuildRuleID        int    `json:"buildRuleId"`
}

// BuildRules BuildRules
type BuildRules struct {
	BuildRules []BuildRule `json:"buildRules"`
}

// BuildImage BuildImage
type BuildImage struct {
	Tag           string `json:"tag"`
	RepoNamespace string `json:"repoNamespace"`
	RepoName      string `json:"repoName"`
}

// Build Build
type Build struct {
	StartTime   int        `json:"startTime"`
	EndTime     int        `json:"endTime"`
	BuildID     string     `json:"buildId"`
	BuildStatus string     `json:"buildStatus"`
	Image       BuildImage `json:"image"`
}

// Builds Builds
type Builds struct {
	Total    int     `json:"total"`
	Page     int     `json:"page"`
	Builds   []Build `json:"builds"`
	PageSize int     `json:"pageSize"`
}
