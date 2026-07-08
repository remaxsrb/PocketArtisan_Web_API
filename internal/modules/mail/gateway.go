package mail

import "context"

// Attachment is a file attached to an email. When ContentID is set, the
// attachment is treated as inline and can be referenced from the HTML body via
// `cid:<ContentID>` (e.g. an embedded logo); otherwise it is a regular
// downloadable attachment.
type Attachment struct {
	Filename    string
	ContentType string
	Content     []byte
	ContentID   string
}

type Message struct {
	To          string
	Subject     string
	HTML        string
	Attachments []Attachment
}

type Service interface {
	Send(ctx context.Context, msg Message) error
}
