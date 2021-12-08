// Copyright © 2021 The Things Industries B.V.

package cmd

import "fmt"

type reportFormat string

var reportFormats = [...]string{
	"json",
	"csv",
	"dot",
	"png",
	"svg",
	"pdf",
	"ps",
}

func newReportFormat(defaultValue string) *reportFormat {
	f := reportFormat(defaultValue)
	return &f
}

func (f reportFormat) String() string {
	return string(f)
}

func (f *reportFormat) Set(s string) error {
	for _, sf := range reportFormats {
		if sf == s {
			*f = reportFormat(s)
			return nil
		}
	}
	return fmt.Errorf("unrecognized format %q", s)
}

func (f reportFormat) Type() string {
	return "reportFormat"
}

func (f reportFormat) ext() string {
	return "." + string(f)
}

func (f reportFormat) isImage() bool {
	switch f {
	case "png", "svg", "pdf", "ps":
		return true
	}
	return false
}
