package connection

import (
	"crypto/rsa"
	"github.com/s-rah/go-ricochet/channels"
	"github.com/s-rah/go-ricochet/policies"
	"github.com/s-rah/go-ricochet/utils"
)

// InboundConnectionHandler is a convieniance wrapper for handling inbound
// connections
type InboundConnectionHandler struct {
	connection *Connection
}

// HandleInboundConnection returns an InboundConnectionHandler given a connection
func HandleInboundConnection(c *Connection) *InboundConnectionHandler {
	ich := new(InboundConnectionHandler)
	ich.connection = c
	return ich
}

// ProcessAuthAsServer blocks until authentication has succeeded, failed, or the
// connection is closed. A non-nil error is returned in all cases other than successful
// and accepted authentication.
//
// ProcessAuthAsServer cannot be called at the same time as any other call to a Process
// function. Another Process function must be called after this function successfully
// returns to continue handling connection events.
//
// The acceptCallback function is called after receiving a valid authentication proof
// with the client's authenticated hostname and public key. acceptCallback must return
// true to accept authentication and allow the connection to continue, and also returns a
// boolean indicating whether the contact is known and recognized. Unknown contacts will
// assume they are required to send a contact request before any other activity.
func (ich *InboundConnectionHandler) ProcessAuthAsServer(privateKey *rsa.PrivateKey, sach func(hostname string, publicKey rsa.PublicKey) (allowed, known bool)) error {

        if privateKey == nil {
                return utils.PrivateKeyNotSetError
        }

	ach := new(AutoConnectionHandler)
	ach.Init(privateKey, ich.connection.RemoteHostname)
	ach.SetServerAuthHandler(sach)

	var authResult channels.AuthChannelResult
	go func() {
		authResult = ach.WaitForAuthenticationEvent()
		ich.connection.Break()
	}()

	policy := policies.UnknownPurposeTimeout
	err := policy.ExecuteAction(func() error {
		return ich.connection.Process(ach)
	})

	if err == nil {
		if authResult.Accepted == true {
			return nil
		}
		return utils.ClientFailedToAuthenticateError
	}

	return err
}
