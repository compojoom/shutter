package keyper

import (
	"crypto/ecdsa"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/brainbot-com/shutter/shuttermint/shmsg"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/types"
)

// RoundInterval is the duration between the start of two consecutive rounds
var RoundInterval time.Duration = time.Duration(5 * time.Second)

// PrivateKeyDeley is the duration between the start of the public key generation and the the start
// of the private key generation for a single round
var PrivateKeyDelay time.Duration = time.Duration(45 * time.Second)

// BatchParams describes the parameters for single Batch identified by the BatchIndex
type BatchParams struct {
	BatchIndex                    uint64
	PublicKeyGenerationStartTime  time.Time
	PrivateKeyGenerationStartTime time.Time
}

// Keyper is used to run the keyper key generation
type Keyper struct {
	SigningKey     *ecdsa.PrivateKey
	ShuttermintURL string
}

// NewBatchParams creates a new BatchParams struct for the given BatchIndex
func NewBatchParams(BatchIndex uint64) BatchParams {
	ts := int64(BatchIndex) * int64(RoundInterval)

	pubstart := time.Unix(ts/int64(time.Second), ts%int64(time.Second))
	privstart := pubstart.Add(PrivateKeyDelay)
	return BatchParams{
		BatchIndex:                    BatchIndex,
		PublicKeyGenerationStartTime:  pubstart,
		PrivateKeyGenerationStartTime: privstart,
	}
}

// NextBatchIndex computes the BatchIndex for the next batch to be started
func NextBatchIndex(t time.Time) uint64 {
	return uint64((t.UnixNano() + int64(RoundInterval) - 1) / int64(RoundInterval))
}

// MessageSender can be used to sign shmsg.Message's and send them to shuttermint
type MessageSender struct {
	rpcclient  client.Client
	signingKey *ecdsa.PrivateKey
}

// NewMessageSender creates a new MessageSender
func NewMessageSender(client client.Client, signingKey *ecdsa.PrivateKey) MessageSender {
	return MessageSender{client, signingKey}
}

// SendMessage signs the given shmsg.Message and sends the message to shuttermint
func (ms MessageSender) SendMessage(msg *shmsg.Message) error {
	signedMessage, err := shmsg.SignMessage(msg, ms.signingKey)
	if err != nil {
		return err
	}
	var tx types.Tx = types.Tx(base64.RawURLEncoding.EncodeToString(signedMessage))
	res, err := ms.rpcclient.BroadcastTxCommit(tx)
	if err != nil {
		return err
	}
	// fmt.Println("broadcast tx", res)
	if res.DeliverTx.Code != 0 {
		return fmt.Errorf("Error in SendMessage: %s", res.DeliverTx.Log)
	}
	return nil
}

// PubkeyGeneratedEvent is generated by shuttermint, when a new public key has been generated.
type PubkeyGeneratedEvent struct {
	BatchIndex uint64
	Pubkey     *ecdsa.PublicKey
}

// PrivkeyGeneratedEvent is generated by shuttermint, when a new private key has been generated
type PrivkeyGeneratedEvent struct {
	BatchIndex uint64
	Privkey    *ecdsa.PrivateKey
}

// BatchConfigEvent is generated by shuttermint, when a new BatchConfg has been added
type BatchConfigEvent struct {
	StartBatchIndex uint64
	Threshhold      uint32
	Keypers         []common.Address
}

// IEvent is an interface for the event types declared above (PubkeyGeneratedEvent,
// PrivkeyGeneratedEvent, BatchConfigEvent)
type IEvent interface {
	IEvent()
}

func (PubkeyGeneratedEvent) IEvent()  {}
func (PrivkeyGeneratedEvent) IEvent() {}
func (BatchConfigEvent) IEvent()      {}
