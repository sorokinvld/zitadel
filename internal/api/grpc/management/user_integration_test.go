//go:build integration

package management_test

import (
	"context"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zitadel/zitadel/internal/integration"
	"github.com/zitadel/zitadel/pkg/grpc/management"
	"github.com/zitadel/zitadel/pkg/grpc/user"
)

var (
	CTX    context.Context
	Tester *integration.Tester
	Client management.ManagementServiceClient
)

func TestMain(m *testing.M) {
	os.Exit(func() int {
		ctx, errCtx, cancel := integration.Contexts(3 * time.Minute)
		defer cancel()

		Tester = integration.NewTester(ctx)
		defer Tester.Done()

		CTX, _ = Tester.WithAuthorization(ctx, integration.OrgOwner), errCtx
		Client = Tester.Client.Mgmt
		return m.Run()
	}())
}

// TestImport_and_Get reproduces https://github.com/zitadel/zitadel/issues/5808
// which led to consistency issues due the call timestamp not being
// updated after a bulk Trigger.
// This test Imports a user and directly tries to Get it, 100 times in a loop.
// When the bug still existed, some (between 1 to 7 out of 100)
// Get calls would return a Not Found error.
func TestImport_and_Get(t *testing.T) {
	const N = 100

	for i := 0; i < N; i++ {
		firstName := strconv.Itoa(i)
		t.Run(firstName, func(t *testing.T) {
			// create unique names.
			lastName := strconv.FormatInt(time.Now().Unix(), 10)
			userName := strings.Join([]string{firstName, lastName}, "_")
			email := strings.Join([]string{userName, "zitadel.com"}, "@")

			res, err := Client.ImportHumanUser(CTX, &management.ImportHumanUserRequest{
				UserName: userName,
				Profile: &management.ImportHumanUserRequest_Profile{
					FirstName:         firstName,
					LastName:          lastName,
					PreferredLanguage: language.Afrikaans.String(),
					Gender:            user.Gender_GENDER_DIVERSE,
				},
				Email: &management.ImportHumanUserRequest_Email{
					Email:           email,
					IsEmailVerified: true,
				},
			})
			require.NoError(t, err)

			_, err = Client.GetUserByID(CTX, &management.GetUserByIDRequest{Id: res.GetUserId()})

			s, ok := status.FromError(err)
			if ok && s != nil && s.Code() == codes.NotFound {
				t.Errorf("iteration %d: user with id %q not found", i, res.GetUserId())
			}
			require.NoError(t, err) // catch and fail on any other error
		})
	}
}
