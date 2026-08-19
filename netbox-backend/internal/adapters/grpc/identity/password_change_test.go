package identity

import (
	"context"
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	identityv1 "netbox-go/gen/go/netbox/identity/v1"
	application "netbox-go/internal/application/identity"
	domain "netbox-go/internal/domain/identity"
)

const (
	i4GRPCCurrentPassword = "grpc-current-password"
	i4GRPCNextPassword    = "grpc-replacement-password"
)

type i4GRPCClock struct{ now time.Time }

func (clock i4GRPCClock) Now() time.Time { return clock.now }

type i4GRPCDigest [sha256.Size]byte

type i4GRPCStore struct {
	application.Store

	users    map[int64]domain.User
	hashes   map[int64]string
	sessions map[i4GRPCDigest]application.SessionRecord

	userLookupErr  error
	transactionErr error
	updateErr      error
	deleteErr      error
	commitErr      error

	transactions int
	userLookups  int
	updates      int
	deletes      int
	creates      int
	tokenLookups int
	tokenTouches int
	tokenID      int64
}

func (store *i4GRPCStore) Transaction(_ context.Context, fn func(application.Store) error) error {
	store.transactions++
	if store.transactionErr != nil {
		return store.transactionErr
	}
	users := i4GRPCCloneUsers(store.users)
	hashes := i4GRPCCloneHashes(store.hashes)
	sessions := i4GRPCCloneSessions(store.sessions)
	if err := fn(store); err != nil {
		store.users = users
		store.hashes = hashes
		store.sessions = sessions
		return err
	}
	if store.commitErr != nil {
		store.users = users
		store.hashes = hashes
		store.sessions = sessions
		return store.commitErr
	}
	return nil
}

func (store *i4GRPCStore) UserByID(_ context.Context, id int64) (domain.User, string, error) {
	store.userLookups++
	if store.userLookupErr != nil {
		return domain.User{}, "", store.userLookupErr
	}
	user, ok := store.users[id]
	if !ok {
		return domain.User{}, "", application.ErrNotFound
	}
	return user, store.hashes[id], nil
}

func (store *i4GRPCStore) UpdatePassword(_ context.Context, id int64, hash string) error {
	store.updates++
	if store.updateErr != nil {
		return store.updateErr
	}
	if _, ok := store.users[id]; !ok {
		return application.ErrNotFound
	}
	store.hashes[id] = hash
	return nil
}

func (store *i4GRPCStore) DeleteSessionsForUser(_ context.Context, userID int64) error {
	store.deletes++
	if store.deleteErr != nil {
		return store.deleteErr
	}
	for key, record := range store.sessions {
		if record.UserID == userID {
			delete(store.sessions, key)
		}
	}
	return nil
}

func (store *i4GRPCStore) CreateSession(_ context.Context, record application.SessionRecord) error {
	store.creates++
	store.sessions[i4GRPCKey(record.SecretHash)] = i4GRPCCloneSession(record)
	return nil
}

func (store *i4GRPCStore) TokenByHash(context.Context, []byte) (application.TokenRecord, domain.User, error) {
	store.tokenLookups++
	return application.TokenRecord{}, domain.User{}, application.ErrNotFound
}

func (store *i4GRPCStore) TouchToken(context.Context, int64, time.Time) error {
	store.tokenTouches++
	return nil
}

func TestGRPCPasswordChangeUsesTokenProvenance(t *testing.T) {
	now := time.Date(2026, 8, 19, 11, 0, 0, 0, time.UTC)
	request := &identityv1.ChangePasswordRequest{
		CurrentPassword: i4GRPCCurrentPassword,
		NewPassword:     i4GRPCNextPassword,
	}

	t.Run("missing principal", func(t *testing.T) {
		store := i4GRPCFixture(t, now)
		response, err := NewServer(application.NewService(store, i4GRPCClock{now: now})).ChangePassword(t.Context(), request)

		require.Nil(t, response)
		require.Equal(t, codes.Unauthenticated, status.Code(err))
		require.Equal(t, "Authentication credentials were not provided.", status.Convert(err).Message())
		require.Zero(t, store.userLookups)
		require.Zero(t, store.transactions)
		require.Zero(t, store.updates)
	})

	for _, validation := range []struct {
		name    string
		request *identityv1.ChangePasswordRequest
	}{
		{
			name: "current password validation",
			request: &identityv1.ChangePasswordRequest{
				CurrentPassword: "incorrect-current-password",
				NewPassword:     i4GRPCNextPassword,
			},
		},
		{
			name: "new password validation",
			request: &identityv1.ChangePasswordRequest{
				CurrentPassword: i4GRPCCurrentPassword,
				NewPassword:     "short",
			},
		},
	} {
		t.Run(validation.name, func(t *testing.T) {
			store := i4GRPCFixture(t, now)
			ctx := domain.WithPrincipal(t.Context(), store.users[41].Principal())
			response, err := NewServer(application.NewService(store, i4GRPCClock{now: now})).ChangePassword(ctx, validation.request)

			require.Nil(t, response)
			require.Equal(t, codes.InvalidArgument, status.Code(err))
			require.Equal(t, "Invalid input.", status.Convert(err).Message())
			require.Zero(t, store.transactions)
			require.Zero(t, store.updates)
			i4GRPCRequirePassword(t, store.hashes[41], i4GRPCCurrentPassword, true)
		})
	}

	t.Run("infrastructure is mapped and concealed", func(t *testing.T) {
		store := i4GRPCFixture(t, now)
		store.userLookupErr = errors.New("password backend unavailable")
		ctx := domain.WithPrincipal(t.Context(), store.users[41].Principal())
		response, err := NewServer(application.NewService(store, i4GRPCClock{now: now})).ChangePassword(ctx, request)

		require.Nil(t, response)
		require.Equal(t, codes.Internal, status.Code(err))
		require.Equal(t, "An internal error occurred.", status.Convert(err).Message())
		require.Zero(t, store.transactions)
		require.Zero(t, store.updates)
	})

	t.Run("success revokes browser sessions without browser result", func(t *testing.T) {
		store := i4GRPCFixture(t, now)
		ctx := domain.WithPrincipal(t.Context(), store.users[41].Principal())
		response, err := NewServer(application.NewService(store, i4GRPCClock{now: now})).ChangePassword(ctx, request)

		require.NoError(t, err)
		require.NotNil(t, response)
		require.Zero(t, response.ProtoReflect().Descriptor().Fields().Len())
		require.Equal(t, 1, store.transactions)
		require.Equal(t, 2, store.userLookups)
		require.Equal(t, 1, store.updates)
		require.Equal(t, 1, store.deletes)
		require.Zero(t, store.creates)
		require.Zero(t, len(store.sessions))
		require.Zero(t, store.tokenLookups)
		require.Zero(t, store.tokenTouches)
		require.Equal(t, int64(73), store.tokenID, "api token state changed")
		i4GRPCRequirePassword(t, store.hashes[41], i4GRPCNextPassword, true)
		i4GRPCRequirePassword(t, store.hashes[41], i4GRPCCurrentPassword, false)
	})

	t.Run("commit failure rolls back all state and returns no response", func(t *testing.T) {
		store := i4GRPCFixture(t, now)
		store.commitErr = errors.New("password transaction commit unavailable")
		ctx := domain.WithPrincipal(t.Context(), store.users[41].Principal())
		response, err := NewServer(application.NewService(store, i4GRPCClock{now: now})).ChangePassword(ctx, request)

		require.Nil(t, response)
		require.Equal(t, codes.Internal, status.Code(err))
		require.Equal(t, "An internal error occurred.", status.Convert(err).Message())
		require.Equal(t, 1, store.transactions)
		require.Equal(t, 1, store.updates)
		require.Equal(t, 1, store.deletes)
		require.Zero(t, store.creates)
		require.Equal(t, 2, len(store.sessions))
		require.Equal(t, int64(73), store.tokenID, "api token state changed")
		i4GRPCRequirePassword(t, store.hashes[41], i4GRPCCurrentPassword, true)
		i4GRPCRequirePassword(t, store.hashes[41], i4GRPCNextPassword, false)
	})
}

func i4GRPCFixture(t *testing.T, now time.Time) *i4GRPCStore {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(i4GRPCCurrentPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatal("could not prepare gRPC password-change fixture")
	}
	store := &i4GRPCStore{
		users: map[int64]domain.User{
			41: {ID: 41, Username: "grpc-user", IsActive: true},
		},
		hashes:   map[int64]string{41: string(hash)},
		sessions: make(map[i4GRPCDigest]application.SessionRecord),
		tokenID:  73,
	}
	for _, secret := range []string{"grpc-session-one", "grpc-session-two"} {
		hash := i4GRPCDigestValue(secret)
		store.sessions[i4GRPCKey(hash)] = application.SessionRecord{
			SecretHash: hash,
			CSRFHash:   i4GRPCDigestValue("fixture-csrf-binding"),
			UserID:     41,
			Created:    now.Add(-time.Hour),
			LastSeen:   now.Add(-time.Hour),
			Expires:    now.Add(time.Hour),
		}
	}
	return store
}

func i4GRPCDigestValue(value string) []byte {
	sum := sha256.Sum256([]byte(value))
	return append([]byte(nil), sum[:]...)
}

func i4GRPCKey(hash []byte) i4GRPCDigest {
	var key i4GRPCDigest
	copy(key[:], hash)
	return key
}

func i4GRPCCloneSession(record application.SessionRecord) application.SessionRecord {
	record.SecretHash = append([]byte(nil), record.SecretHash...)
	record.CSRFHash = append([]byte(nil), record.CSRFHash...)
	return record
}

func i4GRPCCloneSessions(source map[i4GRPCDigest]application.SessionRecord) map[i4GRPCDigest]application.SessionRecord {
	result := make(map[i4GRPCDigest]application.SessionRecord, len(source))
	for key, record := range source {
		result[key] = i4GRPCCloneSession(record)
	}
	return result
}

func i4GRPCCloneUsers(source map[int64]domain.User) map[int64]domain.User {
	result := make(map[int64]domain.User, len(source))
	for id, user := range source {
		result[id] = user
	}
	return result
}

func i4GRPCCloneHashes(source map[int64]string) map[int64]string {
	result := make(map[int64]string, len(source))
	for id, hash := range source {
		result[id] = hash
	}
	return result
}

func i4GRPCRequirePassword(t *testing.T, hash, password string, want bool) {
	t.Helper()
	got := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
	if got != want {
		t.Error("stored password state did not match the expected gRPC outcome")
	}
}
