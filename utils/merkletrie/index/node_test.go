package index

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/grahambrooks/go-git/v5/plumbing"
	"github.com/grahambrooks/go-git/v5/plumbing/format/index"
	"github.com/grahambrooks/go-git/v5/utils/merkletrie"
	"github.com/grahambrooks/go-git/v5/utils/merkletrie/noder"
	"github.com/stretchr/testify/suite"
)

type NoderSuite struct {
	suite.Suite
}

func TestNoderSuite(t *testing.T) {
	suite.Run(t, new(NoderSuite))
}

func (s *NoderSuite) TestDiff() {
	indexA := &index.Index{
		Entries: []*index.Entry{
			{Name: "foo", Hash: plumbing.NewHash("8ab686eafeb1f44702738c8b0f24f2567c36da6d")},
			{Name: "bar/foo", Hash: plumbing.NewHash("8ab686eafeb1f44702738c8b0f24f2567c36da6d")},
			{Name: "bar/qux", Hash: plumbing.NewHash("8ab686eafeb1f44702738c8b0f24f2567c36da6d")},
			{Name: "bar/baz/foo", Hash: plumbing.NewHash("8ab686eafeb1f44702738c8b0f24f2567c36da6d")},
		},
	}

	indexB := &index.Index{
		Entries: []*index.Entry{
			{Name: "foo", Hash: plumbing.NewHash("8ab686eafeb1f44702738c8b0f24f2567c36da6d")},
			{Name: "bar/foo", Hash: plumbing.NewHash("8ab686eafeb1f44702738c8b0f24f2567c36da6d")},
			{Name: "bar/qux", Hash: plumbing.NewHash("8ab686eafeb1f44702738c8b0f24f2567c36da6d")},
			{Name: "bar/baz/foo", Hash: plumbing.NewHash("8ab686eafeb1f44702738c8b0f24f2567c36da6d")},
		},
	}

	ch, err := merkletrie.DiffTree(NewRootNode(indexA), NewRootNode(indexB), isEquals)
	s.NoError(err)
	s.Len(ch, 0)
}

func (s *NoderSuite) TestDiffChange() {
	indexA := &index.Index{
		Entries: []*index.Entry{{
			Name: filepath.Join("bar", "baz", "bar"),
			Hash: plumbing.NewHash("8ab686eafeb1f44702738c8b0f24f2567c36da6d"),
		}},
	}

	indexB := &index.Index{
		Entries: []*index.Entry{{
			Name: filepath.Join("bar", "baz", "foo"),
			Hash: plumbing.NewHash("8ab686eafeb1f44702738c8b0f24f2567c36da6d"),
		}},
	}

	ch, err := merkletrie.DiffTree(NewRootNode(indexA), NewRootNode(indexB), isEquals)
	s.NoError(err)
	s.Len(ch, 2)
}

func (s *NoderSuite) TestDiffDir() {
	indexA := &index.Index{
		Entries: []*index.Entry{{
			Name: "foo",
			Hash: plumbing.NewHash("8ab686eafeb1f44702738c8b0f24f2567c36da6d"),
		}},
	}

	indexB := &index.Index{
		Entries: []*index.Entry{{
			Name: filepath.Join("foo", "bar"),
			Hash: plumbing.NewHash("8ab686eafeb1f44702738c8b0f24f2567c36da6d"),
		}},
	}

	ch, err := merkletrie.DiffTree(NewRootNode(indexA), NewRootNode(indexB), isEquals)
	s.NoError(err)
	s.Len(ch, 2)
}

func (s *NoderSuite) TestDiffSameRoot() {
	indexA := &index.Index{
		Entries: []*index.Entry{
			{Name: "foo.go", Hash: plumbing.NewHash("aab686eafeb1f44702738c8b0f24f2567c36da6d")},
			{Name: "foo/bar", Hash: plumbing.NewHash("8ab686eafeb1f44702738c8b0f24f2567c36da6d")},
		},
	}

	indexB := &index.Index{
		Entries: []*index.Entry{
			{Name: "foo/bar", Hash: plumbing.NewHash("8ab686eafeb1f44702738c8b0f24f2567c36da6d")},
			{Name: "foo.go", Hash: plumbing.NewHash("8ab686eafeb1f44702738c8b0f24f2567c36da6d")},
		},
	}

	ch, err := merkletrie.DiffTree(NewRootNode(indexA), NewRootNode(indexB), isEquals)
	s.NoError(err)
	s.Len(ch, 1)
}

var empty = make([]byte, 24)

func isEquals(a, b noder.Hasher) bool {
	if bytes.Equal(a.Hash(), empty) || bytes.Equal(b.Hash(), empty) {
		return false
	}

	return bytes.Equal(a.Hash(), b.Hash())
}
