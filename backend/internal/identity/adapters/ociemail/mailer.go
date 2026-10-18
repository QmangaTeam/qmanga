package ociemail

import (
	"context"
	"fmt"
	"time"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/common/auth"
	"github.com/oracle/oci-go-sdk/v65/emaildataplane"
)

type submitter interface {
	SubmitEmail(context.Context, emaildataplane.SubmitEmailRequest) (emaildataplane.SubmitEmailResponse, error)
}

type Mailer struct {
	client        submitter
	compartmentID string
	from          string
}

func New(compartmentID, from string) (*Mailer, error) {
	provider, err := auth.InstancePrincipalConfigurationProvider()
	if err != nil {
		return nil, fmt.Errorf("initialize OCI instance principal: %w", err)
	}
	client, err := emaildataplane.NewEmailDPClientWithConfigurationProvider(provider)
	if err != nil {
		return nil, fmt.Errorf("initialize OCI email client: %w", err)
	}
	return NewWithClient(&client, compartmentID, from), nil
}

func NewWithClient(client submitter, compartmentID, from string) *Mailer {
	return &Mailer{client: client, compartmentID: compartmentID, from: from}
}

func (mailer *Mailer) Send(ctx context.Context, to, subject, body string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	request := emaildataplane.SubmitEmailRequest{SubmitEmailDetails: emaildataplane.SubmitEmailDetails{
		Sender: &emaildataplane.Sender{
			CompartmentId: common.String(mailer.compartmentID),
			SenderAddress: &emaildataplane.EmailAddress{Email: common.String(mailer.from)},
		},
		Recipients: &emaildataplane.Recipients{To: []emaildataplane.EmailAddress{{Email: common.String(to)}}},
		Subject:    common.String(subject),
		BodyText:   common.String(body),
	}}
	if _, err := mailer.client.SubmitEmail(ctx, request); err != nil {
		return fmt.Errorf("submit OCI email: %w", err)
	}
	return nil
}
