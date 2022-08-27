package repomanager

import (
	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/pkg/errors"
	"sort"
	"strings"
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
		return nil, errors.Wrap(err, "open git repo")
	}

	return &Manager{
		repo: r,
	}, nil
}

func (m *Manager) GetTagsAll() ([]*plumbing.Reference, error) {
	tagReferences, err := m.repo.Tags()
	if err != nil {
		return nil, errors.Wrap(err, "get repo tags")
	}

	var tags []*plumbing.Reference
	err = tagReferences.ForEach(func(t *plumbing.Reference) error {
		tags = append(tags, t)
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "get repo tags")
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

// GetTagsSemver returns only semver tags
func (m *Manager) GetTagsSemver() ([]SemverTag, error) {
	references, err := m.GetTagsAll()
	if err != nil {
		return nil, errors.Wrap(err, "get all tags")
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
		return nil, errors.Wrap(err, "get semver tags")
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
		return nil, errors.Wrap(ErrNotFound, "has no semver tags")
	}

	return &maxTag, nil
}

// GetTagsSemverTopN return top n semver tags
func (m *Manager) GetTagsSemverTopN(n int) ([]SemverTag, error) {
	tags, err := m.GetTagsSemver()
	if err != nil {
		return nil, errors.Wrap(err, "get semver tags")
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

// IncrementSemverTag will increment max semver tag and write tag to repo
func (m *Manager) IncrementSemverTag(c Component) (*SemverTag, *SemverTag, error) {
	maxTag, err := m.GetTagsSemverMax()
	switch errors.Cause(err) {
	default:
		return nil, nil, errors.Wrap(err, "get max semver tag")
	case ErrNotFound:
		if errors.Is(err, ErrNotFound) {
			maxTag = &SemverTag{
				Version: *semver.MustParse("v0.0.0"),
				Ref:     nil,
			}
		}
	case nil:
	}

	newVersion := maxTag.Version
	switch c {
	default:
		return nil, nil, errors.Wrap(err, "unknown component")
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
		return nil, nil, errors.Wrap(err, "get repo head")
	}

	ref, err := m.repo.CreateTag(versionStr, head.Hash(), nil)
	if err != nil {
		return nil, nil, errors.Wrap(err, "create tag")
	}

	return maxTag, &SemverTag{
		Version: newVersion,
		Ref:     ref,
	}, nil
}
