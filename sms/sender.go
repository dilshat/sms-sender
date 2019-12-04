package sms

import (
	"github.com/cskr/pubsub"
	"github.com/dilshat/sms-sender/log"
	"time"
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
	Send(id uint32, sender, phone, text string)
	BindSubmitSmResponseHandler(handler func(id, status uint32, smscId uint64))
	BindDeliverSmHandler(handler func(smscId uint64, status string))
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

func (s *sender) BindSubmitSmResponseHandler(handler func(id, status uint32, smscId uint64)) {
	s.smppClient.BindSubmitSmResponseHandler(handler)
}

func (s *sender) BindDeliverSmHandler(handler func(smscId uint64, status string)) {
	s.smppClient.BindDeliverSmHandler(handler)
}

func (s *sender) Send(id uint32, sender, phone, text string) {
	s.ps.Pub(sms{Id: id, Sender: sender, Phone: phone, Text: text}, OUT)
}
func (s *sender) ReadPackets() {
	for {
		if s.smppClient.IsConnected() {
			err := s.smppClient.ReadPacket()
			log.ErrIfErr("", err)
		}
	}
}

func (s *sender) CheckConnection() {
	for {
		if !s.smppClient.IsConnected() {
			err := s.smppClient.Reconnect()
			log.ErrIfErr("", err)
		}
		time.Sleep(time.Second)
	}
}

func (s *sender) processOutgoing() {
	for {
		if s.smppClient.IsConnected() {
			select {
			case val, ok := <-s.out:
				if ok {
					sms := val.(sms)
					err := s.smppClient.SendMessage(sms.Id, sms.Sender, sms.Phone, sms.Text)
					log.ErrIfErr("", err)
				} else {
					//channel closed, should not arrive here
					return
				}
			default:
				//no new messages, no-op
				continue
			}
		} else {
			time.Sleep(time.Second)
		}
	}
}
