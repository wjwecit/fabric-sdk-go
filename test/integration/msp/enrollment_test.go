/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp/api"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
)

const (
	IdentityTypeUser = "User"
	sdkConfigFile    = "../../fixtures/config/config_test.yaml"
)

func TestRegisterEnroll(t *testing.T) {

	configProvider := config.FromFile(sdkConfigFile)

	// Instantiate the SDK
	sdk, err := fabsdk.New(configProvider)
	if err != nil {
		t.Fatalf("SDK init failed: %v", err)
	}

	ctxProvider := sdk.Context()

	// Get the MSP.
	// Without WithOrg option, uses default client organization.
	msp, err := msp.New(ctxProvider)
	if err != nil {
		t.Fatalf("failed to create CA client: %v", err)
	}

	// As this integration test spawns a fresh CA instance,
	// we have to enroll the CA registrar first. Otherwise,
	// CA operations that require the registrar's identity
	// will be rejected by the CA.
	registrarEnrollID, registrarEnrollSecret := getRegistrarEnrollmentCredentials(t, sdk.Config())
	err = msp.Enroll(registrarEnrollID, registrarEnrollSecret)
	if err != nil {
		t.Fatalf("Enroll failed: %v", err)
	}

	// The enrollment process generates a new private key and
	// enrollment certificate for the user. The private key
	// is stored in the SDK crypto provider's key store, while the
	// enrollment certificate is stored in the SKD's user store
	// (state store). The CAClient will lookup the
	// registrar's identity information in these stores.

	// Generate a random user name
	userName := integration.GenerateRandomID()

	// Register the new user
	enrollmentSecret, err := msp.Register(&api.RegistrationRequest{
		Name: userName,
		Type: IdentityTypeUser,
		// Affiliation is mandatory. "org1" and "org2" are hardcoded as CA defaults
		// See https://github.com/hyperledger/fabric-ca/blob/release/cmd/fabric-ca-server/config.go
		Affiliation: "org2",
	})
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	// Enroll the new user
	err = msp.Enroll(userName, enrollmentSecret)
	if err != nil {
		t.Fatalf("Enroll failed: %v", err)
	}

	// Get the new user's signing identity
	_, err = msp.GetSigningIdentity(userName)
	if err != nil {
		t.Fatalf("GetSigningIdentity failed: %v", err)
	}

	// Get the new user's full information
	_, err = msp.GetUser(userName)
	if err != nil {
		t.Fatalf("GetSigningIdentity failed: %v", err)
	}

}

func getRegistrarEnrollmentCredentials(t *testing.T, config core.Config) (string, string) {

	clientConfig, err := config.Client()
	if err != nil {
		t.Fatalf("config.MSP() failed: %v", err)
	}

	myOrg := clientConfig.Organization

	caConfig, err := config.CAConfig(myOrg)
	if err != nil {
		t.Fatalf("CAConfig failed: %v", err)
	}

	return caConfig.Registrar.EnrollID, caConfig.Registrar.EnrollSecret
}
