package service

import (
	"github.com/asdine/storm/v3/codec/json"
	"github.com/dilshat/sms-sender/model"
	"github.com/dilshat/sms-sender/service/dto"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

const (
	STATUS_STORE_DAYS int    = 7
	MSG_MAX_LEN              = 300
	ID                uint32 = 123
	SENDER                   = "Awesome"
	TEXT                     = "What is up?"
	PHONE                    = "996ZZZXXXXXX"
	PHONE2                   = "996YYYAABBCC"
	JSON_MESSAGE             = `{"id":123,"sender":"Awesome","text":"What is up?","statuses":[{"phone":"996ZZZXXXXXX","status":"DELIVRD"},{"phone":"996YYYAABBCC","status":"ACCEPTD"}]}`
	JSON_RECIPIENT           = `{"id":123,"sender":"Awesome","text":"What is up?","statuses":[{"phone":"996ZZZXXXXXX","status":"DELIVRD"}]}`
	PHONE_MASK               = "996\\w{9}"
)

var (
	submitStatusUpdated     bool
	deliverStatusUpdated    bool
	cleanupMessagesCalled   bool
	cleanupRecipientsCalled bool
)

type mockMessageDao struct {
}

func (m mockMessageDao) RemoveOlderThanDays(days int) error {
	cleanupMessagesCalled = true
	return nil
}

func (m mockMessageDao) Create(text, sender string) (uint32, error) {
	return 1, nil
}

func (m mockMessageDao) GetOneById(id uint32) (model.Message, error) {
	return model.Message{
		Id:        ID,
		Text:      TEXT,
		Sender:    SENDER,
		CreatedAt: time.Time{},
	}, nil
}

func (m mockMessageDao) GetAll() ([]model.Message, error) {
	return nil, nil
}

type mockRecipientDao struct {
}

func (m mockRecipientDao) RemoveOlderThanDays(days int) error {
	cleanupRecipientsCalled = true
	return nil
}

func (m mockRecipientDao) Create(messageId uint32, phone string) (uint32, error) {
	return 2, nil
}

func (m mockRecipientDao) UpdateSubmitStatus(id uint32, deliverId string, status string) error {
	submitStatusUpdated = true
	return nil
}

func (m mockRecipientDao) UpdateDeliverStatus(deliverId string, status string) (uint32, string, error) {
	deliverStatusUpdated = true
	return 0, "", nil
}

func (m mockRecipientDao) GetOneByMessageIdAndPhone(messageId uint32, phone string) (model.Recipient, error) {
	return model.Recipient{
		Id:        1,
		MessageId: ID,
		Phone:     PHONE,
		Status:    "DELIVRD",
		DeliverId: "321",
	}, nil
}

func (m mockRecipientDao) GetAllByMessageId(messageId uint32) ([]model.Recipient, error) {
	return []model.Recipient{
		{
			Id:        1,
			MessageId: ID,
			Phone:     PHONE,
			Status:    "DELIVRD",
			DeliverId: "321",
			CreatedAt: time.Now(),
		},
		{
			Id:        2,
			MessageId: ID,
			Phone:     PHONE2,
			Status:    "ACCEPTD",
			DeliverId: "567",
			CreatedAt: time.Now(),
		},
	}, nil
}

func (m mockRecipientDao) GetAll() ([]model.Recipient, error) {
	return nil, nil
}

type mockSender struct {
}

func (m mockSender) Start() error {
	return nil
}

func (m mockSender) BindSubmitSmResponseHandler(handler func(id, status uint32, smscId string)) {
}

func (m mockSender) BindDeliverSmHandler(handler func(smscId string, status string)) {
}

func (m mockSender) Send(id uint32, sender, phone, text string) error {
	return nil
}

func TestService_SendMessage(t *testing.T) {
	service := NewService(mockSender{}, mockMessageDao{}, mockRecipientDao{}, STATUS_STORE_DAYS, MSG_MAX_LEN, "", PHONE_MASK)

	id, err := service.SendMessage(dto.Message{
		Sender: SENDER,
		Text:   TEXT,
		Phones: []string{PHONE},
	})

	require.NoError(t, err)
	require.NotEmpty(t, id)
	require.True(t, id.Id > 0)

	time.Sleep(time.Millisecond * 100)

	require.True(t, cleanupMessagesCalled)
	require.True(t, cleanupRecipientsCalled)
}

func TestService_CheckStatusOfMessage(t *testing.T) {
	service := NewService(mockSender{}, mockMessageDao{}, mockRecipientDao{}, STATUS_STORE_DAYS, MSG_MAX_LEN, "", PHONE_MASK)

	status, err := service.CheckStatusOfMessage(ID)

	require.NoError(t, err)
	require.NotEmpty(t, status)

	b, err := json.Codec.Marshal(status)
	if err != nil {
		t.Error(err)
	}

	require.JSONEq(t, JSON_MESSAGE, string(b))
}

func TestService_CheckStatusOfRecipient(t *testing.T) {
	service := NewService(mockSender{}, mockMessageDao{}, mockRecipientDao{}, STATUS_STORE_DAYS, MSG_MAX_LEN, "", PHONE_MASK)

	status, err := service.CheckStatusOfRecipient(ID, PHONE)

	require.NoError(t, err)
	require.NotEmpty(t, status)

	b, err := json.Codec.Marshal(status)
	if err != nil {
		t.Error(err)
	}

	require.JSONEq(t, JSON_RECIPIENT, string(b))
}

func TestImp_HandleSubmitSmResp(t *testing.T) {
	impl := &service{
		sender:       mockSender{},
		messageDao:   mockMessageDao{},
		recipientDao: mockRecipientDao{},
	}

	impl.HandleSubmitSmResp(ID, 0, "123")

	require.True(t, submitStatusUpdated)
}

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func TestImp_HandleDeliverSm(t *testing.T) {

	client := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			//Body:       ioutil.NopCloser(bytes.NewBufferString(`OK`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	impl := &service{
		sender:       mockSender{},
		messageDao:   mockMessageDao{},
		recipientDao: mockRecipientDao{},
		httpClient:   client,
		webhook:      "http://www.kg",
	}

	impl.HandleDeliverSm("123", "status")

	require.True(t, deliverStatusUpdated)
}
