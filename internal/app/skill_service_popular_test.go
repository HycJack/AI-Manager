package app

import (
	"testing"
)

// leaderboardFixture mirrors how skills.sh actually ships the data: Next.js
// puts the leaderboard inside an escaped RSC string, so in the raw HTML the
// quotes arrive as \" .
const leaderboardFixture = `<script>self.__next_f.push([1,"4e:[\"$\",\"$L55\",null,{\"initialSkills\":[{\"source\":\"vercel-labs/skills\",\"skillId\":\"find-skills\",\"name\":\"find-skills\",\"installs\":3394646,\"weeklyInstalls\":[107969,101120,96861,93130,100221,103058,102072,101832],\"isOfficial\":true},{\"source\":\"a/b\",\"skillId\":\"weird-skill\",\"name\":\"weird \\\"name\\\" [x]\",\"installs\":10}]"}])</script>`

func TestExtractSkillsShLeaderboard(t *testing.T) {
	skills, err := extractSkillsShLeaderboard(leaderboardFixture)
	if err != nil {
		t.Fatalf("extractSkillsShLeaderboard: %v", err)
	}
	if len(skills) != 2 {
		t.Fatalf("got %d entries, want 2", len(skills))
	}

	top := skills[0]
	if top.Source != "vercel-labs/skills" || top.SkillID != "find-skills" {
		t.Errorf("entry 0 = %+v", top)
	}
	if top.Installs != 3394646 {
		t.Errorf("installs = %d, want 3394646", top.Installs)
	}
	if !top.IsOfficial {
		t.Error("entry 0 should be official")
	}
	if len(top.WeeklyInstalls) != 8 {
		t.Errorf("weeklyInstalls len = %d, want 8", len(top.WeeklyInstalls))
	}

	// A name containing both a quote and a bracket must round-trip: the
	// bracket scanner has to stay inside JSON strings.
	if got := skills[1].Name; got != `weird "name" [x]` {
		t.Errorf("entry 1 name = %q, want %q", got, `weird "name" [x]`)
	}
}

func TestExtractSkillsShLeaderboard_NotFound(t *testing.T) {
	if _, err := extractSkillsShLeaderboard("<html>no leaderboard here</html>"); err == nil {
		t.Error("expected an error when the key is absent")
	}
}

func TestExtractSkillsShLeaderboard_Unterminated(t *testing.T) {
	if _, err := extractSkillsShLeaderboard(`<script>self.__next_f.push([1,"{\"initialSkills\":[{\"source\":\"a/b\"}")</script>`); err == nil {
		t.Error("expected an error for an unterminated array")
	}
}

func TestRankPopularSkills(t *testing.T) {
	in := []skillsShLeaderboardSkill{
		{Source: "owner/mid", SkillID: "mid", Name: "mid", Installs: 500, WeeklyInstalls: []int{10, 20}},
		{Source: "owner/top", SkillID: "top", Name: "top", Installs: 900, WeeklyInstalls: []int{5, 5}, IsOfficial: true},
		{Source: "owner/top", SkillID: "top", Name: "top-duplicate", Installs: 850}, // duplicate skillId, drops out
		{Source: "owner/low", SkillID: "low", Name: "low", Installs: 100},
		{Source: "no-slash", SkillID: "bad", Name: "bad", Installs: 2000},   // skipped: not owner/repo
		{Source: "owner/", SkillID: "", Name: "empty-slug", Installs: 3000}, // skipped: empty skillId
	}

	out := rankPopularSkills(in, 2)

	if len(out) != 2 {
		t.Fatalf("got %d rows, want 2", len(out))
	}
	if out[0].Key != "top" || out[0].Rank != 1 || out[0].Installs != 900 {
		t.Errorf("row 1 = %+v", out[0])
	}
	if !out[0].IsOfficial {
		t.Error("row 1 should keep isOfficial")
	}
	if out[0].GithubURL != "https://github.com/owner/top" {
		t.Errorf("githubUrl = %q", out[0].GithubURL)
	}
	if out[0].RepoOwner != "owner" || out[0].RepoName != "top" {
		t.Errorf("owner/repo = %q/%q", out[0].RepoOwner, out[0].RepoName)
	}
	if out[0].WeeklyActivity != 10 {
		t.Errorf("weeklyActivity = %d, want 10", out[0].WeeklyActivity)
	}

	if out[1].Key != "mid" || out[1].Rank != 2 {
		t.Errorf("row 2 = %+v", out[1])
	}
	if out[1].WeeklyActivity != 30 {
		t.Errorf("row 2 weeklyActivity = %d, want 30", out[1].WeeklyActivity)
	}
}

func TestRankPopularSkills_LimitZero(t *testing.T) {
	out := rankPopularSkills([]skillsShLeaderboardSkill{
		{Source: "a/b", SkillID: "b", Name: "b", Installs: 1},
	}, 0)
	if len(out) != 0 {
		t.Errorf("limit 0 should yield no rows, got %d", len(out))
	}
}
