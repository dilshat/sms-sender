package sms

import (
	"errors"
	"time"

	"github.com/cskr/pubsub"
	"go.uber.org/zap"
)

const (
	OUT = "out"
)

type sms struct {
	Id     uint32
	Sender string
	Text   string
	Phone  string
}

type Response struct {
}

type Sender interface {
	Start() error
	Send(id uint32, sender, phone, text string) error
	BindSubmitSmResponseHandler(handler func(id, status uint32, smscId string))
	BindDeliverSmHandler(handler func(smscId string, status string))
}

type sender struct {
	smppClient SmppClient
	ps         *pubsub.PubSub
	out        chan interface{}
}

func NewSender(smppClient SmppClient) Sender {
	ps := pubsub.New(100)
	return &sender{smppClient: smppClient, ps: ps, out: ps.Sub(OUT)}
}

func (s *sender) Start() error {

	err := s.smppClient.Connect()
	if err != nil {
		return err
	}

	go s.ReadPackets()
	go s.CheckConnection()
	go s.processOutgoing()

	return nil
}

func (s *sender) BindSubmitSmResponseHandler(handler func(id, status uint32, smscId string)) {
	s.smppClient.BindSubmitSmResponseHandler(handler)
}

func (s *sender) BindDeliverSmHandler(handler func(smscId string, status string)) {
	s.smppClient.BindDeliverSmHandler(handler)
}

func (s *sender) Send(id uint32, sender, phone, text string) error {
	if !s.smppClient.IsConnected() {
		return errors.New("Not connected to SMSC")
	}

	s.ps.Pub(sms{Id: id, Sender: sender, Phone: phone, Text: text}, OUT)

	return nil
}
func (s *sender) ReadPackets() {
	for {
		if s.smppClient.IsConnected() {
			err := s.smppClient.ReadPacket()
			if err != nil {
				zap.L().Error("Error reading packets", zap.Error(err))
			}
		} else {
			time.Sleep(time.Second)
		}
	}
}

func (s *sender) CheckConnection() {
	for {
		if !s.smppClient.IsConnected() {
			err := s.smppClient.Reconnect()
			if err != nil {
				zap.L().Error("Error reconnecting", zap.Error(err))
			}
		}
		time.Sleep(time.Second * 5)
	}
}

func (s *sender) processOutgoing() {
	sleepDuration := time.Microsecond * 500
	for {
		if s.smppClient.IsConnected() {
			select {
			case val, ok := <-s.out:
				if ok {
					sms := val.(sms)
					err := s.smppClient.SendMessage(sms.Id, sms.Sender, sms.Phone, sms.Text)
					if err != nil {
						zap.L().Error("Error sending message", zap.Error(err))
					}
					//sleep to avoid sending messages without pauses
					time.Sleep(sleepDuration)
				} else {
					//channel closed, should not arrive here
					return
				}
			default:
				//no new messages, no-op
				time.Sleep(time.Second)
			}
		} else {
			time.Sleep(time.Second)
		}
	}
}
