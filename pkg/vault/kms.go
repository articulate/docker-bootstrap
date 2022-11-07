package vault

import (
	"context"
	"encoding/base64"

	"github.com/aws/aws-sdk-go-v2/service/kms"
)

type KMSDecryptAPI interface {
	Decrypt(context.Context, *kms.DecryptInput, ...func(*kms.Options)) (*kms.DecryptOutput, error)
}

func DecodeToken(ctx context.Context, api KMSDecryptAPI, token string) (string, error) {
	blob, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", err
	}

	result, err := api.Decrypt(ctx, &kms.DecryptInput{
		CiphertextBlob: blob,
	})
	if err != nil {
		return "", err
	}

	return string(result.Plaintext), nil
}
