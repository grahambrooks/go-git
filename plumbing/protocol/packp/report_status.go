package packp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/grahambrooks/go-git/v5/plumbing"
	"github.com/grahambrooks/go-git/v5/plumbing/format/pktline"
)

const (
	ok = "ok"
)

// ReportStatus is a report status message, as used in the git-receive-pack
// process whenever the 'report-status' capability is negotiated.
type ReportStatus struct {
	UnpackStatus    string
	CommandStatuses []*CommandStatus
}

// NewReportStatus creates a new ReportStatus message.
func NewReportStatus() *ReportStatus {
	return &ReportStatus{}
}

// Error returns the first error if any.
func (s *ReportStatus) Error() error {
	if s.UnpackStatus != ok {
		return fmt.Errorf("unpack error: %s", s.UnpackStatus)
	}

	for _, s := range s.CommandStatuses {
		if err := s.Error(); err != nil {
			return err
		}
	}

	return nil
}

// Encode writes the report status to a writer.
func (s *ReportStatus) Encode(w io.Writer) error {
	if _, err := pktline.Writef(w, "unpack %s\n", s.UnpackStatus); err != nil {
		return err
	}

	for _, cs := range s.CommandStatuses {
		if err := cs.encode(w); err != nil {
			return err
		}
	}

	return pktline.WriteFlush(w)
}

// Decode reads from the given reader and decodes a report-status message. It
// does not read more input than what is needed to fill the report status.
func (s *ReportStatus) Decode(r io.Reader) error {
	b, err := s.scanFirstLine(r)
	if err != nil {
		return err
	}

	if err := s.decodeReportStatus(b); err != nil {
		return err
	}

	var l int
	flushed := false
	for {
		l, b, err = pktline.ReadLine(r)
		if err != nil {
			break
		}

		if l == pktline.Flush {
			flushed = true
			break
		}

		if err := s.decodeCommandStatus(b); err != nil {
			return err
		}
	}

	if !flushed {
		return fmt.Errorf("missing flush")
	}

	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	return nil
}

func (s *ReportStatus) scanFirstLine(r io.Reader) ([]byte, error) {
	_, p, err := pktline.ReadLine(r)
	if errors.Is(err, io.EOF) {
		return p, io.ErrUnexpectedEOF
	}
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (s *ReportStatus) decodeReportStatus(b []byte) error {
	if isFlush(b) {
		return fmt.Errorf("premature flush")
	}

	b = bytes.TrimSuffix(b, eol)

	line := string(b)
	fields := strings.SplitN(line, " ", 2)
	if len(fields) != 2 || fields[0] != "unpack" {
		return fmt.Errorf("malformed unpack status: %s", line)
	}

	s.UnpackStatus = fields[1]
	return nil
}

func (s *ReportStatus) decodeCommandStatus(b []byte) error {
	b = bytes.TrimSuffix(b, eol)

	line := string(b)
	fields := strings.SplitN(line, " ", 3)
	status := ok
	if len(fields) == 3 && fields[0] == "ng" {
		status = fields[2]
	} else if len(fields) != 2 || fields[0] != "ok" {
		return fmt.Errorf("malformed command status: %s", line)
	}

	cs := &CommandStatus{
		ReferenceName: plumbing.ReferenceName(fields[1]),
		Status:        status,
	}
	s.CommandStatuses = append(s.CommandStatuses, cs)
	return nil
}

// CommandStatus is the status of a reference in a report status.
// See ReportStatus struct.
type CommandStatus struct {
	ReferenceName plumbing.ReferenceName
	Status        string
}

// Error returns the error, if any.
func (s *CommandStatus) Error() error {
	if s.Status == ok {
		return nil
	}

	return fmt.Errorf("command error on %s: %s",
		s.ReferenceName.String(), s.Status)
}

func (s *CommandStatus) encode(w io.Writer) error {
	if s.Error() == nil {
		_, err := pktline.Writef(w, "ok %s\n", s.ReferenceName.String())
		return err
	}

	_, err := pktline.Writef(w, "ng %s %s\n", s.ReferenceName.String(), s.Status)
	return err
}
