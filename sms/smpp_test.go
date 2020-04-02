package sms

import (
	"errors"
	"reflect"
	"runtime"
	"testing"

	"github.com/Dilshat/smpp34"
	"github.com/dchest/uniuri"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

const (
	SEQ    uint32 = 1
	SENDER        = "sender"
	PHONE         = "996777123456"
)

var (
	unbound           bool
	closed            bool
	nextId            uint32
	submitCount       int
	deliverSmRespSent bool
)

func TestSmppClient_ReadPacket(t *testing.T) {
	smppClnt := smppClient{transceiver: transceiverWrapperMock{err: errors.New("blablabla")}}

	err := smppClnt.ReadPacket()

	require.Error(t, err)
	require.False(t, smppClnt.IsConnected())

	//SUBMIT_SM_RESP
	pdu := mockPdu{header: &smpp34.Header{Id: smpp34.SUBMIT_SM_RESP, Sequence: SEQ, Status: 0},
		field: mockField{str: "1203837180"}}
	smppClnt = smppClient{transceiver: transceiverWrapperMock{pdu: pdu}}
	f := func(id, status uint32, smscId uint64) {}
	smppClnt.BindSubmitSmResponseHandler(f)

	err = smppClnt.ReadPacket()

	require.NoError(t, err)

	//DELIVER_SM
	deliverSmRespSent = false
	pdu = mockPdu{header: &smpp34.Header{Id: smpp34.DELIVER_SM},
		field: mockField{str: "id:1203837180  sub:001 dlvrd:1  submit date:1911251537 done date:1911251537 stat:DELIVRD err:000  TEXT:a message space. What is up bro?"}}
	smppClnt = smppClient{transceiver: transceiverWrapperMock{pdu: pdu}}
	f2 := func(smscId uint64, status string) {}
	smppClnt.BindDeliverSmHandler(f2)

	err = smppClnt.ReadPacket()

	require.NoError(t, err)
	require.True(t, deliverSmRespSent)
}

func TestSmppClient_SendMessage(t *testing.T) {
	nextId = 0
	submitCount = 0
	smppClnt := smppClient{transceiver: transceiverWrapperMock{}, rateLimiter: rate.NewLimiter(rate.Limit(1), 1)}

	err := smppClnt.SendMessage(SEQ, SENDER, PHONE, uniuri.NewLen(10))

	require.NoError(t, err)
	require.Equal(t, SEQ, nextId)
	require.Equal(t, 1, submitCount)

	submitCount = 0
	err = smppClnt.SendMessage(SEQ, SENDER, PHONE, uniuri.NewLen(400))

	require.NoError(t, err)
	require.Equal(t, 3, submitCount)

	submitCount = 0
	err = smppClnt.SendMessage(SEQ, SENDER, PHONE, uniuri.NewLen(100)+"привет")

	require.NoError(t, err)
	require.Equal(t, 2, submitCount)
}

func TestSmppClient_Reconnect(t *testing.T) {
	unbound = false
	closed = false
	smppClnt := smppClient{connected: 0, transceiverFactory: transceiverWrapperFactoryMock{}}

	err := smppClnt.Reconnect()

	require.NoError(t, err)
	require.True(t, smppClnt.IsConnected())
}

func TestSmppClient_Connect(t *testing.T) {
	smppClnt := smppClient{transceiverFactory: transceiverWrapperFactoryMock{}}

	err := smppClnt.Connect()

	require.NoError(t, err)
	require.True(t, smppClnt.IsConnected())

	smppClnt.transceiverFactory = transceiverWrapperFactoryMock{err: errors.New("blablabla")}

	err = smppClnt.Connect()

	require.Error(t, err)
	require.False(t, smppClnt.IsConnected())
}

func TestSmppClient_IsConnected(t *testing.T) {
	smppClnt := smppClient{connected: 1}

	require.True(t, smppClnt.IsConnected())

	smppClnt.connected = 0

	require.False(t, smppClnt.IsConnected())
}

func TestSmppClient_Disconnect(t *testing.T) {
	smppClnt := smppClient{transceiver: transceiverWrapperMock{}}

	smppClnt.Disconnect()

	require.True(t, unbound)
	require.True(t, closed)
}

func TestSmppClient_BindDeliverSmHandler(t *testing.T) {
	smppClnt := smppClient{}
	f := func(smscId uint64, status string) {}

	smppClnt.BindDeliverSmHandler(f)

	require.Equal(t, runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(), runtime.FuncForPC(reflect.ValueOf(smppClnt.deliverHandler).Pointer()).Name())
}

func TestSmppClient_BindSubmitSmResponseHandler(t *testing.T) {
	smppClnt := smppClient{}
	f := func(id, status uint32, smscId uint64) {}

	smppClnt.BindSubmitSmResponseHandler(f)

	require.Equal(t, runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(), runtime.FuncForPC(reflect.ValueOf(smppClnt.submitSmHandler).Pointer()).Name())
}

//----------------------mocks------------

type mockField struct {
	str string
}

func (m mockField) Length() interface{} {
	panic("implement me")
}

func (m mockField) Value() interface{} {
	panic("implement me")
}

func (m mockField) String() string {
	return m.str
}

func (m mockField) ByteArray() []byte {
	panic("implement me")
}

type mockPdu struct {
	header *smpp34.Header
	field  mockField
}

func (m mockPdu) Fields() map[string]smpp34.Field {
	panic("implement me")
}

func (m mockPdu) MandatoryFieldsList() []string {
	panic("implement me")
}

func (m mockPdu) GetField(string) smpp34.Field {
	return m.field
}

func (m mockPdu) GetHeader() *smpp34.Header {
	return m.header
}

func (m mockPdu) TLVFields() map[uint16]*smpp34.TLVField {
	panic("implement me")
}

func (m mockPdu) Writer() []byte {
	panic("implement me")
}

func (m mockPdu) SetField(f string, v interface{}) error {
	panic("implement me")
}

func (m mockPdu) SetTLVField(t, l int, v []byte) error {
	panic("implement me")
}

func (m mockPdu) SetSeqNum(uint32) {
	panic("implement me")
}

func (m mockPdu) Ok() bool {
	panic("implement me")
}

type transceiverWrapperFactoryMock struct {
	err error
}

func (t transceiverWrapperFactoryMock) GetTransceiver(host string, port int, eli int, bindParams smpp34.Params) (TransceiverWrapper, error) {
	return transceiverWrapperMock{}, t.err
}

type transceiverWrapperMock struct {
	pdu smpp34.Pdu
	err error
}

func (t transceiverWrapperMock) Unbind() error {
	unbound = true
	return nil
}

func (t transceiverWrapperMock) Close() {
	closed = true
}

func (t transceiverWrapperMock) SetNextId(id uint32) {
	nextId = id
}

func (t transceiverWrapperMock) Read() (smpp34.Pdu, error) {
	return t.pdu, t.err
}

func (t transceiverWrapperMock) SubmitSmEncoded(sourceAddr, destinationAddr string, shortMessage []byte, params *smpp34.Params) (seq uint32, err error) {
	submitCount++
	return SEQ, nil
}

func (t transceiverWrapperMock) DeliverSmResp(seq uint32, status smpp34.CMDStatus) error {
	deliverSmRespSent = true
	return nil
}
