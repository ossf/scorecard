package command

import (
	"context"
	"fmt"

	containeranalysis "cloud.google.com/go/containeranalysis/apiv1"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/genproto/googleapis/grafeas/v1"
)

const NOTE_ID = "ossf-scorecard"

func CreateAttestation() {
	ctx := context.Background()
	credentials, err := google.FindDefaultCredentials(ctx, compute.ComputeScope)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf(credentials.ProjectID)

	c, err := containeranalysis.NewClient(ctx)
	if err != nil {
		// TODO: handle error
	}

	g := c.GetGrafeasClient()
	req := grafeas.CreateNoteRequest{
		Parent: fmt.Sprintf("projects/%s", credentials.ProjectID),
		NoteId: NOTE_ID,
	}

	note, err := g.CreateNote(ctx, req, opts)

	if err != nil {
		// TODO: handle error
	}
}
