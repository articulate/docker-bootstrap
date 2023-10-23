package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type kmsMock struct {
	mock.Mock
}

func (m *kmsMock) Decrypt(
	ctx context.Context,
	input *kms.DecryptInput,
	_ ...func(*kms.Options),
) (*kms.DecryptOutput, error) {
	args := m.Called(ctx, input)

	return &kms.DecryptOutput{
		Plaintext: []byte(args.String(0)),
	}, args.Error(1) //nolint:wrapcheck
}

func TestDecodeToken(t *testing.T) {
	m := new(kmsMock)

	t.Run("invalid base64", func(t *testing.T) {
		token, err := decodeToken(context.TODO(), m, "dfk0dEJ#0(#$@)(")
		assert.Equal(t, "", token)
		require.Error(t, err)
	})

	t.Run("decrypt", func(t *testing.T) {
		m.On("Decrypt", context.TODO(), &kms.DecryptInput{
			CiphertextBlob: []byte("test"),
		}).Return("my-decrypted-token", nil)

		token, err := decodeToken(context.TODO(), m, "dGVzdA==")
		require.NoError(t, err)
		assert.Equal(t, "my-decrypted-token", token)
	})

	t.Run("kms error", func(t *testing.T) {
		m.On("Decrypt", context.TODO(), &kms.DecryptInput{
			CiphertextBlob: []byte("foobar"),
		}).Return("no", fmt.Errorf("kms error")) //nolint:goerr113

		token, err := decodeToken(context.TODO(), m, "Zm9vYmFy")
		assert.Equal(t, "", token)
		require.EqualError(t, err, "could not decrypt token: kms error")
	})

	m.AssertExpectations(t)
}
