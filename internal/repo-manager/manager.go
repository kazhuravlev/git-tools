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
	references, err := m.GetTagsAll()
	if err != nil {
		return nil, fmt.Errorf("get all tags: %w", err)
	}

	res := make([]SemverTag, 0, len(references))
	for i := range references {
		ref := references[i]

		tagName := ref.Name().Short()

		version, err := semver.NewVersion(tagName)
		if err != nil {
			continue
		}

		res = append(res, SemverTag{
			Version: *version,
			Ref:     ref,
		})
	}

	return res, nil
}

// GetTagsSemverMax return tag that point to max semver version
func (m *Manager) GetTagsSemverMax() (*SemverTag, error) {
	tags, err := m.GetTagsSemver()
	if err != nil {
		return nil, fmt.Errorf("get semver tags: %w", err)
	}

	maxTag := SemverTag{
		Version: *semver.MustParse("v0.0.0"),
		Ref:     nil,
	}
	var found bool
	for i := range tags {
		if tags[i].Version.GreaterThan(&maxTag.Version) {
			found = true
			maxTag = tags[i]
		}
	}

	if !found {
		return nil, fmt.Errorf("has no semver tags: %w", ErrNotFound)
	}

	return &maxTag, nil
}

// GetTagsSemverTopN return top n semver tags
func (m *Manager) GetTagsSemverTopN(n int) ([]SemverTag, error) {
	tags, err := m.GetTagsSemver()
	if err != nil {
		return nil, fmt.Errorf("get semver tags: %w", err)
	}

	sort.SliceStable(tags, func(i, j int) bool {
		return tags[i].Version.LessThan(&tags[j].Version)
	})

	res := make([]SemverTag, 0, n)
	for i := range tags {
		if i == n {
			break
		}

		res = append(res, tags[i])
	}

	return res, nil
}

// GetCurrentTagSemver return a tag if that is presented for current commit. It will ignore all non-semver tags.
func (m *Manager) GetCurrentTagSemver() (optional.Val[SemverTag], error) {
	head, err := m.repo.Head()
	if err != nil {
		return optional.Empty[SemverTag](), fmt.Errorf("get repo head: %w", err)
	}

	tagReferences, err := m.repo.Tags()
	if err != nil {
		return optional.Empty[SemverTag](), fmt.Errorf("get repo tags: %w", err)
	}

	var tag optional.Val[SemverTag]
	{
		err := tagReferences.ForEach(func(t *plumbing.Reference) error {
			if t.Hash() == head.Hash() {
				version, err := semver.NewVersion(t.Name().Short())
				if err != nil {
					return nil
				}

				tag = optional.New(SemverTag{
					Version: *version,
					Ref:     t,
				})

				return storer.ErrStop
			}

			return nil
		})
		if err != nil {
			return optional.Empty[SemverTag](), fmt.Errorf("get repo tags: %w", err)
		}
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
