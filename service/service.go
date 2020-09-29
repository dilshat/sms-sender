package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dilshat/sms-sender/dao"
	"github.com/dilshat/sms-sender/model"
	"github.com/dilshat/sms-sender/service/dto"
	"github.com/dilshat/sms-sender/sms"
	"github.com/dilshat/sms-sender/util"
	"go.uber.org/zap"
)

type InvalidPayloadErr struct {
	message string
}

func (e *InvalidPayloadErr) Error() string {
	return e.message
}

func NewInvalidPayloadError(msg string) *InvalidPayloadErr {
	return &InvalidPayloadErr{message: msg}
}

type Service interface {
	SendMessage(message dto.Message) (dto.Id, error)
	CheckStatusOfMessage(id uint32) (dto.MessageStatus, error)
	CheckStatusOfRecipient(id uint32, phone string) (dto.MessageStatus, error)
}
type service struct {
	sender          sms.Sender
	messageDao      dao.MessageDao
	recipientDao    dao.RecipientDao
	httpClient      *http.Client
	statusStoreDays int
	messageMaxLen   int
	webhook         string
	phoneRx         *regexp.Regexp
}

func NewService(sender sms.Sender, messageDao dao.MessageDao, recipientDao dao.RecipientDao, statusStoreDays, messageMaxLen int, webhook, phoneMask string) Service {
	service := &service{
		sender:          sender,
		messageDao:      messageDao,
		recipientDao:    recipientDao,
		statusStoreDays: statusStoreDays,
		messageMaxLen:   messageMaxLen,
		webhook:         webhook,
		phoneRx:         regexp.MustCompile(phoneMask),
		httpClient:      &http.Client{Timeout: 10 * time.Second},
	}

	sender.BindDeliverSmHandler(service.HandleDeliverSm)
	sender.BindSubmitSmResponseHandler(service.HandleSubmitSmResp)

	go service.CleanupDb()

	return service
}

func (s service) CleanupDb() {
	for {
		err := s.messageDao.RemoveOlderThanDays(s.statusStoreDays)
		if err != nil {
			zap.L().Warn("Error cleaning up messages", zap.Error(err))
		}
		err = s.recipientDao.RemoveOlderThanDays(s.statusStoreDays)
		if err != nil {
			zap.L().Warn("Error cleaning up recipients", zap.Error(err))
		}
		time.Sleep(time.Hour)
	}
}

func (s service) HandleSubmitSmResp(id, status uint32, smscId string) {
	smStatus := model.SUBMIT_OK
	if status != 0 {
		smStatus = model.SUBMIT_FAIL
	}
	err := s.recipientDao.UpdateSubmitStatus(id, smscId, smStatus)
	if err != nil {
		zap.L().Error("Error updating submit status", zap.Error(err))
	}
}

func (s service) HandleDeliverSm(smscId string, status string) {
	msgId, phone, err := s.recipientDao.UpdateDeliverStatus(smscId, status)
	if err != nil {
		zap.L().Error("Error updating delivery status", zap.Error(err))
		return
	}

	if util.IsBlank(s.webhook) {
		return
	}

	msgStatus, err := s.CheckStatusOfRecipient(msgId, phone)
	if err != nil {
		zap.L().Error("Error checking recipient message status", zap.Error(err))
		return
	}

	msgStatusBytes, err := json.Marshal(msgStatus)
	if err != nil {
		zap.L().Error("Error checking recipient message status", zap.Error(err))
		return
	}

	client := &http.Client{Timeout: time.Second * 10}

	req, err := http.NewRequest("POST", s.webhook, bytes.NewBuffer(msgStatusBytes))
	if err != nil {
		zap.L().Error("Error calling web hook", zap.Error(err))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		zap.L().Error("Error calling web hook", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	if !(resp.StatusCode >= 200 && resp.StatusCode <= 202) {
		zap.L().Warn("Webhook returned unexpected status", zap.String("status", resp.Status))
	}
}

func (s service) SendMessage(message dto.Message) (dto.Id, error) {

	//overall message validation
	if strings.TrimSpace(message.Text) == "" || strings.TrimSpace(message.Sender) == "" || len(message.Phones) == 0 {
		return dto.Id{}, NewInvalidPayloadError("Invalid message ")
	}

	//check phone format
	for _, phone := range message.Phones {
		if !s.phoneRx.MatchString(phone) {
			return dto.Id{}, NewInvalidPayloadError("Invalid phone " + phone)
		}
	}

	//check max length of sms
	if len([]rune(message.Text)) > s.messageMaxLen {
		return dto.Id{}, NewInvalidPayloadError("Message too long. Must be <= " + strconv.Itoa(s.messageMaxLen) + " symbols in length")
	}

	msgId, err := s.messageDao.Create(message.Text, message.Sender)
	if err != nil {
		return dto.Id{}, err
	}

	//remove duplicates
	uniquePhones := make(map[string]bool)
	for _, phone := range message.Phones {
		uniquePhones[phone] = true
	}

	for phone := range uniquePhones {
		id, err := s.recipientDao.Create(msgId, phone)
		if err != nil {
			return dto.Id{}, err
		}

		err = s.sender.Send(id, message.Sender, phone, message.Text)
		if err != nil {
			return dto.Id{}, err
		}
	}

	return dto.Id{Id: msgId}, nil
}

func (s service) CheckStatusOfMessage(id uint32) (dto.MessageStatus, error) {
	msg, err := s.messageDao.GetOneById(id)
	if err != nil {
		return dto.MessageStatus{}, err
	}
	recipients, err := s.recipientDao.GetAllByMessageId(msg.Id)
	if err != nil {
		return dto.MessageStatus{}, err
	}

	status := dto.MessageStatus{
		Id:     msg.Id,
		Sender: msg.Sender,
		Text:   msg.Text,
	}
	recipientStatuses := []dto.RecipientStatus{}
	for _, rs := range recipients {
		recipientStatuses = append(recipientStatuses, dto.RecipientStatus{
			Phone:  rs.Phone,
			Status: rs.Status,
		})
	}
	status.Statuses = recipientStatuses

	return status, nil
}

func (s service) CheckStatusOfRecipient(id uint32, phone string) (dto.MessageStatus, error) {
	msg, err := s.messageDao.GetOneById(id)
	if err != nil {
		return dto.MessageStatus{}, err
	}
	recipient, err := s.recipientDao.GetOneByMessageIdAndPhone(msg.Id, phone)
	if err != nil {
		return dto.MessageStatus{}, err
	}

	status := dto.MessageStatus{
		Id:     msg.Id,
		Sender: msg.Sender,
		Text:   msg.Text,
	}
	recipientStatuses := []dto.RecipientStatus{
		{
			Phone:  recipient.Phone,
			Status: recipient.Status,
		},
	}
	status.Statuses = recipientStatuses

	return status, nil
}
