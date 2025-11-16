package v1

import "github.com/squ1ky/pr-manager/internal/entity"

func toPullRequestDTO(pr *entity.PullRequest, reviewers []string) PullRequestDTO {
	dto := PullRequestDTO{
		PullRequestID:     pr.ID,
		PullRequestName:   pr.Name,
		AuthorID:          pr.AuthorID,
		Status:            string(pr.Status),
		AssignedReviewers: reviewers,
	}

	dto.MergedAt = pr.MergedAt

	return dto
}

func toPullRequestShortDTO(pr *entity.PullRequest) PullRequestShortDTO {
	return PullRequestShortDTO{
		PullRequestID:   pr.ID,
		PullRequestName: pr.Name,
		AuthorID:        pr.AuthorID,
		Status:          string(pr.Status),
	}
}

func toPullRequestShortList(prs []entity.PullRequest) []PullRequestShortDTO {
	res := make([]PullRequestShortDTO, 0, len(prs))
	for _, pr := range prs {
		res = append(res, toPullRequestShortDTO(&pr))
	}
	return res
}

func toTeamMemberDTO(u *entity.User) TeamMemberDTO {
	return TeamMemberDTO{
		UserID:   u.ID,
		Username: u.Username,
		IsActive: &u.IsActive,
	}
}

func teamMembersToUsers(members []TeamMemberDTO) []entity.User {
	res := make([]entity.User, 0, len(members))
	for _, m := range members {
		res = append(res, entity.User{
			ID:       m.UserID,
			Username: m.Username,
			IsActive: *m.IsActive,
		})
	}
	return res
}

func toTeamDTO(team *entity.Team, members []entity.User) TeamDTO {
	res := TeamDTO{
		TeamName: team.Name,
		Members:  make([]TeamMemberDTO, 0, len(members)),
	}

	for _, u := range members {
		res.Members = append(res.Members, toTeamMemberDTO(&u))
	}

	return res
}

func toUserDTO(u *entity.User) UserDTO {
	var teamName string
	if u.TeamName != nil {
		teamName = *u.TeamName
	}

	return UserDTO{
		UserID:   u.ID,
		Username: u.Username,
		TeamName: teamName,
		IsActive: u.IsActive,
	}
}
