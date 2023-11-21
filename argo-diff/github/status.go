package github

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v56/github"
	"github.com/rs/zerolog/log"
)

var (
	statusClient     *github.Client
	statusContextStr = "argo-diff"
)

const StatusPending = "pending"
const StatusSuccess = "success"
const StatusFailure = "failure"
const StatusError = "error"

func init() {
	githubPAT := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
	if githubPAT == "" {
		log.Fatal().Msg("Cannot create github client - GITHUB_PERSONAL_ACCESS_TOKEN is empty")
	} else {
		statusClient = github.NewClient(nil).WithAuthToken(githubPAT)
	}
	statusContextEnv := os.Getenv("GITHUB_STATUS_CONTEXT_STR")
	if statusContextEnv != "" {
		statusContextStr = statusContextEnv
	}
}

func Status(ctx context.Context, status, description, repoOwner, repoName, commitSha string) error {
	if status != StatusPending && status != StatusSuccess && status != StatusFailure && status != StatusError {
		log.Fatal().Msgf("Cannot create github status with status string '%s'", status)
		return fmt.Errorf("unknown status string '%s'", status)
	}
	// TODO add support for AvatarURL ?
	// TODO add support for TargetURL ?
	repoStatus := &github.RepoStatus{
		State:       &status,
		Description: &description,
		Context:     github.String(statusContextStr),
	}
	_, resp, err := statusClient.Repositories.CreateStatus(ctx, repoOwner, repoName, commitSha, repoStatus)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to create repo status %s/%s@%s: %s %s '%s'", repoOwner, repoName, commitSha, statusContextStr, status, description)
		return err
	}
	log.Info().Msgf("%s - repo status %s/%s@%s: %s %s '%s'", resp.Status, repoOwner, repoName, commitSha, statusContextStr, status, description)
	return nil
}
