package ociemail

import (
	"context"
	"testing"

	"github.com/oracle/oci-go-sdk/v65/emaildataplane"
)

type captureClient struct {
	request emaildataplane.SubmitEmailRequest
}

func (client *captureClient) SubmitEmail(_ context.Context, request emaildataplane.SubmitEmailRequest) (emaildataplane.SubmitEmailResponse, error) {
	client.request = request
	return emaildataplane.SubmitEmailResponse{}, nil
}

func TestSendSubmitsTextEmailFromApprovedSender(testingTarget *testing.T) {
	client := &captureClient{}
	mailer := NewWithClient(client, "ocid1.tenancy.oc1..example", "no-reply@maqsatto.dev")
	if err := mailer.Send(context.Background(), "reader@example.com", "Verify Qmanga", "Open the link"); err != nil {
		testingTarget.Fatal(err)
	}
	request := client.request.SubmitEmailDetails
	if request.Sender == nil || request.Sender.CompartmentId == nil || *request.Sender.CompartmentId != "ocid1.tenancy.oc1..example" {
		testingTarget.Fatalf("wrong email compartment: %+v", request.Sender)
	}
	if request.Sender.SenderAddress == nil || request.Sender.SenderAddress.Email == nil || *request.Sender.SenderAddress.Email != "no-reply@maqsatto.dev" {
		testingTarget.Fatalf("wrong sender: %+v", request.Sender)
	}
	if request.Recipients == nil || len(request.Recipients.To) != 1 || request.Recipients.To[0].Email == nil || *request.Recipients.To[0].Email != "reader@example.com" {
		testingTarget.Fatalf("wrong recipient: %+v", request.Recipients)
	}
	if request.Subject == nil || *request.Subject != "Verify Qmanga" || request.BodyText == nil || *request.BodyText != "Open the link" {
		testingTarget.Fatalf("wrong message: %+v", request)
	}
}
