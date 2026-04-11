package repomanager

import (
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/kazhuravlev/optional"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type Component string

const (
	ComponentMajor Component = "major"
	ComponentMinor Component = "minor"
	ComponentPatch Component = "patch"
)

var ErrNotFound = errors.New("not found")

type Manager struct {
	repo *git.Repository
}

func New(path string) (*Manager, error) {
	r, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("open git repo: %w", err)
	}

	return &Manager{
		repo: r,
	}, nil
}

func (m *Manager) GetTagsAll() ([]*plumbing.Reference, error) {
	tagReferences, err := m.repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("get repo tags: %w", err)
	}

	var tags []*plumbing.Reference
	err = tagReferences.ForEach(func(t *plumbing.Reference) error {
		tags = append(tags, t)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("get repo tags: %w", err)
	}

	return tags, nil
}

func (m *Manager) forEachSemverTag(fn func(SemverTag) error) error {
	tagReferences, err := m.repo.Tags()
	if err != nil {
		return fmt.Errorf("get repo tags: %w", err)
	}

	err = tagReferences.ForEach(func(t *plumbing.Reference) error {
		version, err := semver.NewVersion(t.Name().Short())
		if err != nil {
			return nil
		}

		return fn(SemverTag{
			Version: *version,
			Ref:     t,
		})
	})
	if err != nil {
		return fmt.Errorf("iterate semver tags: %w", err)
	}

	return nil
}

type SemverTag struct {
	Version semver.Version
	Ref     *plumbing.Reference
}

func (t SemverTag) HasPrefixV() bool {
	return strings.HasPrefix(t.Version.Original(), "v")
}

func (t SemverTag) TagName() string {
	return t.Version.Original()
}

func (t SemverTag) CommitHash() string {
	return t.Ref.Hash().String()
}

// GetTagsSemver returns only semver tags
func (m *Manager) GetTagsSemver() ([]SemverTag, error) {
	res := make([]SemverTag, 0)
	err := m.forEachSemverTag(func(tag SemverTag) error {
		res = append(res, tag)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("get semver tags: %w", err)
	}

	return res, nil
}

// GetTagsSemverMax return tag that point to max semver version
func (m *Manager) GetTagsSemverMax() (*SemverTag, error) {
	var maxTag SemverTag
	var found bool
	err := m.forEachSemverTag(func(tag SemverTag) error {
		if !found || tag.Version.GreaterThan(&maxTag.Version) {
			found = true
			maxTag = tag
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("get semver tags: %w", err)
	}

	if !found {
		return nil, fmt.Errorf("has no semver tags: %w", ErrNotFound)
	}

	return &maxTag, nil
}

// GetTagsSemverTopN return top n semver tags
func (m *Manager) GetTagsSemverTopN(n int) ([]SemverTag, error) {
	if n <= 0 {
		return nil, nil
	}

	top := make([]SemverTag, 0, n)
	err := m.forEachSemverTag(func(tag SemverTag) error {
		if len(top) < n {
			top = append(top, tag)
			sort.Slice(top, func(i, j int) bool {
				return top[i].Version.LessThan(&top[j].Version)
			})
			return nil
		}

		if !top[0].Version.LessThan(&tag.Version) {
			return nil
		}

		top[0] = tag
		sort.Slice(top, func(i, j int) bool {
			return top[i].Version.LessThan(&top[j].Version)
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("get semver tags: %w", err)
	}

	sort.Slice(top, func(i, j int) bool {
		return top[i].Version.GreaterThan(&top[j].Version)
	})

	return top, nil
}

// GetCurrentTagSemver return a tag if that is presented for current commit. It will ignore all non-semver tags.
func (m *Manager) GetCurrentTagSemver() (optional.Val[SemverTag], error) {
	head, err := m.repo.Head()
	if err != nil {
		return optional.Empty[SemverTag](), fmt.Errorf("get repo head: %w", err)
	}

	var tag optional.Val[SemverTag]
	err = m.forEachSemverTag(func(candidate SemverTag) error {
		if candidate.Ref.Hash() != head.Hash() {
			return nil
		}

		tag = optional.New(candidate)
		return storer.ErrStop
	})
	if err != nil && !errors.Is(err, storer.ErrStop) {
		return optional.Empty[SemverTag](), fmt.Errorf("get repo tags: %w", err)
	}

	return tag, nil
}

// IncrementSemverTag will increment max semver tag and write tag to repo
func (m *Manager) IncrementSemverTag(c Component) (*SemverTag, *SemverTag, error) {
	maxTag, err := m.GetTagsSemverMax()
	switch {
	case errors.Is(err, ErrNotFound):
		maxTag = &SemverTag{
			Version: *semver.MustParse("v0.0.0"),
			Ref:     nil,
		}
	case err != nil:
		return nil, nil, fmt.Errorf("get max semver tag: %w", err)
	case err == nil:
	}

	newVersion := maxTag.Version
	switch c {
	default:
	case ComponentMajor:
		newVersion = newVersion.IncMajor()
	case ComponentMinor:
		newVersion = newVersion.IncMinor()
	case ComponentPatch:
		newVersion = newVersion.IncPatch()
	}

	versionStr := newVersion.String()
	if maxTag.HasPrefixV() {
		versionStr = "v" + versionStr
	}

	head, err := m.repo.Head()
	if err != nil {
		return nil, nil, fmt.Errorf("get repo head: %w", err)
	}

	ref, err := m.repo.CreateTag(versionStr, head.Hash(), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create tag: %w", err)
	}

	return maxTag, &SemverTag{
		Version: newVersion,
		Ref:     ref,
	}, nil
}

func (m *Manager) GetCurrentBranch() (string, error) {
	ref, err := m.repo.Head()
	if err != nil {
		return "", fmt.Errorf("get repo head: %w", err)
	}

	if ref.Name().IsBranch() {
		branchName := ref.Name().Short()

		return branchName, nil
	}

	return "", fmt.Errorf("HEAD is not pointing to a branch")
}

type Author struct {
	Name  string
	Email string
	Count int
}

func (m *Manager) GetAuthors() ([]Author, error) {
	commitIter, err := m.repo.CommitObjects()
	if err != nil {
		return nil, fmt.Errorf("get commit objects: %w", err)
	}

	authorMap := make(map[string]*Author)

	err = commitIter.ForEach(func(c *object.Commit) error {
		email := c.Author.Email
		if _, exists := authorMap[email]; !exists {
			authorMap[email] = &Author{
				Name:  c.Author.Name,
				Email: email,
				Count: 0,
			}
		}
		authorMap[email].Count++

		// Update name if it's different (use the most recent name)
		if authorMap[email].Name != c.Author.Name && c.Author.When.After(c.Committer.When) {
			authorMap[email].Name = c.Author.Name
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("iterate commits: %w", err)
	}

	// Convert map to slice
	authors := make([]Author, 0, len(authorMap))
	for _, author := range authorMap {
		authors = append(authors, *author)
	}

	// Sort by commit count (descending)
	sort.Slice(authors, func(i, j int) bool {
		return authors[i].Count > authors[j].Count
	})

	return authors, nil
}
