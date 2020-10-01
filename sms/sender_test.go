package sms

import (
	"testing"
	"time"

	"github.com/cskr/pubsub"
	"github.com/stretchr/testify/require"
)

var (
	submitHandlerBound  bool
	deliverHandlerBound bool
	connectCount        int
	packetsCount        int
	messageSent         bool
)

type mockSmppClient struct {
	connnected bool
	panic      bool
}

func (m mockSmppClient) Connect() error {
	return nil
}

func (m mockSmppClient) IsConnected() bool {
	return m.connnected
}

func (m mockSmppClient) SendMessage(id uint32, from, phone, text string) error {
	messageSent = true
	return nil
}

func (m mockSmppClient) Disconnect() {
	panic("implement me")
}

func (m mockSmppClient) Reconnect() error {
	if m.panic {
		connectCount++
		if connectCount > 1 {
			panic("break loop")
		}
	}
	return nil
}

func (m mockSmppClient) BindSubmitSmResponseHandler(handler func(id, status uint32, smscId string)) {
	submitHandlerBound = true
}

func (m mockSmppClient) BindDeliverSmHandler(handler func(smscId string, status string)) {
	deliverHandlerBound = true
}

func (m mockSmppClient) ReadPacket() error {
	if m.panic {
		packetsCount++
		if packetsCount > 1 {
			panic("")
		}
	}

	return nil
}

func TestSender_Start(t *testing.T) {
	ps := pubsub.New(1)
	sender := sender{ps: ps, out: ps.Sub(OUT), smppClient: &mockSmppClient{connnected: true}}
	sender.Send(123, "sender", "phone", "text")

	err := sender.Start()
	time.Sleep(time.Second * 2)

	require.NoError(t, err)
	require.True(t, messageSent)
}

func TestSender_Send(t *testing.T) {
	ps := pubsub.New(0)
	sender := sender{ps: ps, out: ps.Sub(OUT), smppClient: mockSmppClient{connnected: true}}

	sender.Send(123, "sender", "phone", "text")

	val, ok := <-sender.out

	require.True(t, ok)
	require.IsType(t, sms{}, val)
}

func TestSender_ReadPackets(t *testing.T) {
	defer func() {
		recover()
		require.Equal(t, 2, packetsCount)
	}()

	sender := sender{smppClient: mockSmppClient{connnected: true, panic: true}}

	sender.ReadPackets()
}

func TestSender_CheckConnection(t *testing.T) {
	defer func() {
		recover()
		require.Equal(t, 2, connectCount)
	}()

	sender := sender{smppClient: mockSmppClient{panic: true}}

	sender.CheckConnection()
}

func TestSender_BindDeliverSmHandler(t *testing.T) {
	sender := NewSender(mockSmppClient{})

	sender.BindDeliverSmHandler(func(smscId string, status string) {
	})

	require.True(t, deliverHandlerBound)
}

func TestSender_BindSubmitSmResponseHandler(t *testing.T) {
	sender := NewSender(mockSmppClient{})

	sender.BindSubmitSmResponseHandler(func(id, status uint32, smscId string) {

	})

	require.True(t, submitHandlerBound)
}
