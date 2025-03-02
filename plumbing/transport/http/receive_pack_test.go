package http

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/grahambrooks/go-git/v5/plumbing/transport/test"

	fixtures "github.com/go-git/go-git-fixtures/v4"
)

func TestReceivePackSuite(t *testing.T) {
	suite.Run(t, new(ReceivePackSuite))
}

type ReceivePackSuite struct {
	rps test.ReceivePackSuite
	BaseSuite
}

func (s *ReceivePackSuite) SetupTest() {
	s.BaseSuite.SetupTest()

	s.rps.SetS(s)
	s.rps.Client = DefaultClient
	s.rps.Endpoint = s.prepareRepository(fixtures.Basic().One(), "basic.git")
	s.rps.EmptyEndpoint = s.prepareRepository(fixtures.ByTag("empty").One(), "empty.git")
	s.rps.NonExistentEndpoint = s.newEndpoint("non-existent.git")
}
