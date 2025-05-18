// This package was primarily developed using LLM models and should NOT be considered reliable.
// The purpose of this package is to provide functionality for creating, updating, and deleting feed records on Bluesky, as no suitable tools were found for this purpose.
// If a reliable tool becomes available that can perform these operations, this package will be deprecated and the discovered tool will be referenced in the project instead.

package manage

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aykhans/bsky-feedgen/pkg/utils"
	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/xrpc"
)

const (
	DefaultPDSHost = "https://bsky.social"
)

func NewClient(pdsHost *string) *xrpc.Client {
	if pdsHost == nil {
		pdsHost = utils.ToPtr(DefaultPDSHost)
	}

	return &xrpc.Client{
		Host: *pdsHost,
	}
}

func NewClientWithAuth(ctx context.Context, client *xrpc.Client, identifier, password string) (*xrpc.Client, error) {
	if client == nil {
		client = NewClient(nil)
	}

	auth, err := atproto.ServerCreateSession(ctx, client, &atproto.ServerCreateSession_Input{
		Identifier: identifier,
		Password:   password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create auth session: %v", err)
	}

	client.Auth = &xrpc.AuthInfo{
		AccessJwt:  auth.AccessJwt,
		RefreshJwt: auth.RefreshJwt,
		Did:        auth.Did,
		Handle:     auth.Handle,
	}

	return client, nil
}

func uploadBlob(ctx context.Context, clientWithAuth *xrpc.Client, avatarPath string) (*atproto.RepoUploadBlob_Output, error) {
	if clientWithAuth == nil {
		return nil, fmt.Errorf("client can't be nil")
	}
	if clientWithAuth.Auth == nil {
		return nil, fmt.Errorf("client auth can't be nil")
	}

	avatarFile, err := os.Open(avatarPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open avatar file: %v", err)
	}
	defer func() { _ = avatarFile.Close() }()

	uploadResp, err := atproto.RepoUploadBlob(ctx, clientWithAuth, avatarFile)
	if err != nil {
		return nil, fmt.Errorf("failed to upload avatar: %v", err)
	}

	return uploadResp, nil
}

func GetFeedGenerator(ctx context.Context, clientWithAuth *xrpc.Client, keyName string) (*atproto.RepoGetRecord_Output, error) {
	if clientWithAuth == nil {
		return nil, fmt.Errorf("client can't be nil")
	}
	if clientWithAuth.Auth == nil {
		return nil, fmt.Errorf("client auth can't be nil")
	}

	record, err := atproto.RepoGetRecord(
		ctx,
		clientWithAuth,
		"",
		"app.bsky.feed.generator",
		clientWithAuth.Auth.Did,
		keyName,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get axisting feed generator: %v", err)
	}

	return record, nil
}

func CreateFeedGenerator(
	ctx context.Context,
	clientWithAuth *xrpc.Client,
	displayName string,
	description *string,
	avatarPath *string,
	did string,
	keyName string,
) error {
	if clientWithAuth == nil {
		return fmt.Errorf("client can't be nil")
	}
	if clientWithAuth.Auth == nil {
		return fmt.Errorf("client auth can't be nil")
	}

	var avatarBlob *lexutil.LexBlob
	if avatarPath != nil {
		uploadResp, err := uploadBlob(ctx, clientWithAuth, *avatarPath)
		if err != nil {
			return err
		}

		avatarBlob = uploadResp.Blob
	}

	record := bsky.FeedGenerator{
		DisplayName: displayName,
		Description: description,
		Avatar:      avatarBlob,
		Did:         did,
		CreatedAt:   time.Now().Format(time.RFC3339Nano),
	}

	wrappedRecord := &lexutil.LexiconTypeDecoder{
		Val: &record,
	}

	_, err := atproto.RepoCreateRecord(ctx, clientWithAuth, &atproto.RepoCreateRecord_Input{
		Collection: "app.bsky.feed.generator",
		Repo:       clientWithAuth.Auth.Did, // Your DID (the one creating the record)
		Record:     wrappedRecord,
		Rkey:       &keyName,
	})
	if err != nil {
		return fmt.Errorf("failed to create feed generator: %v", err)
	}

	return nil
}

func UpdateFeedGenerator(
	ctx context.Context,
	clientWithAuth *xrpc.Client,
	displayName *string,
	description *string,
	avatarPath *string,
	did *string,
	keyName string,
) error {
	if clientWithAuth == nil {
		return fmt.Errorf("client can't be nil")
	}
	if clientWithAuth.Auth == nil {
		return fmt.Errorf("client auth can't be nil")
	}

	existingRecord, err := GetFeedGenerator(ctx, clientWithAuth, keyName)
	if err != nil {
		return fmt.Errorf("failed to get axisting feed generator: %v", err)
	}

	if existingRecord != nil && existingRecord.Value != nil {
		if existingFeedgen, ok := existingRecord.Value.Val.(*bsky.FeedGenerator); ok {
			if avatarPath != nil {
				uploadResp, err := uploadBlob(ctx, clientWithAuth, *avatarPath)
				if err != nil {
					return err
				}

				existingFeedgen.Avatar = uploadResp.Blob
			}

			if displayName != nil {
				existingFeedgen.DisplayName = *displayName
			}

			if description != nil {
				existingFeedgen.Description = description
			}

			if did != nil {
				existingFeedgen.Did = *did
			}

			wrappedExistingFeedgen := &lexutil.LexiconTypeDecoder{
				Val: &bsky.FeedGenerator{
					DisplayName: existingFeedgen.DisplayName,
					Description: existingFeedgen.Description,
					Did:         existingFeedgen.Did,
					Avatar:      existingFeedgen.Avatar,
					CreatedAt:   existingFeedgen.CreatedAt,
				},
			}

			_, err := atproto.RepoPutRecord(ctx, clientWithAuth, &atproto.RepoPutRecord_Input{
				Collection: "app.bsky.feed.generator",
				Repo:       clientWithAuth.Auth.Did, // Your DID
				Rkey:       keyName,                 // The Rkey of the record to update
				Record:     wrappedExistingFeedgen,
				SwapRecord: existingRecord.Cid,
			})
			if err != nil {
				return fmt.Errorf("failed to update feed generator: %v", err)
			}
		} else {
			return fmt.Errorf("feed generator not found")
		}
	}

	return nil
}

func DeleteFeedGenerator(
	ctx context.Context,
	clientWithAuth *xrpc.Client,
	keyName string,
) error {
	if clientWithAuth == nil {
		return fmt.Errorf("client can't be nil")
	}
	if clientWithAuth.Auth == nil {
		return fmt.Errorf("client auth can't be nil")
	}

	f, err := atproto.RepoDeleteRecord(ctx, clientWithAuth, &atproto.RepoDeleteRecord_Input{
		Collection: "app.bsky.feed.generator",
		Repo:       clientWithAuth.Auth.Did,
		Rkey:       keyName,
	})
	if err != nil {
		return fmt.Errorf("failed to delete feed generator: %v", err)
	}
	if f.Commit == nil {
		return fmt.Errorf("feed generator not found")
	}

	return nil
}
