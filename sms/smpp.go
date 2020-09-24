package sms

import (
	"context"
	"crypto/rand"
	"math"
	"regexp"
	"strconv"
	"sync/atomic"

	smpp "github.com/Dilshat/smpp34"
	"github.com/Dilshat/smpp34/gsmutil"
	"github.com/dilshat/sms-sender/log"
	"github.com/dilshat/sms-sender/util"
	"golang.org/x/time/rate"
)

var (
	dlvRctRx = *regexp.MustCompile(`(?s)id:(.+?) .* stat:([A-Z]+)`)
)

type RateLimiter interface {
	// Wait blocks until the limiter permits an event to happen.
	Wait(ctx context.Context) error
}

type TransceiverWrapper interface {
	Unbind() error
	Close()
	SetNextId(id uint32)
	Read() (smpp.Pdu, error)
	SubmitSmEncoded(sourceAddr, destinationAddr string, shortMessage []byte, params *smpp.Params) (seq uint32, err error)
	DeliverSmResp(seq uint32, status smpp.CMDStatus) error
}

type TransceiverWrapperFactory interface {
	GetTransceiver(host string, port int, eli int, bindParams smpp.Params) (TransceiverWrapper, error)
}

type transceiverWrapperFactory struct {
}

type transceiverWrapper struct {
	tr *smpp.Transceiver
}

func (t *transceiverWrapper) Unbind() error {
	return t.tr.Unbind()
}

func (t *transceiverWrapper) Close() {
	t.tr.Close()
}

func (t *transceiverWrapper) SetNextId(id uint32) {
	t.tr.NewSeqNumFunc = func() uint32 {
		return id
	}
}

func (t *transceiverWrapper) Read() (smpp.Pdu, error) {
	return t.tr.Read()
}

func (t *transceiverWrapper) SubmitSmEncoded(sourceAddr, destinationAddr string, shortMessage []byte, params *smpp.Params) (seq uint32, err error) {
	return t.tr.SubmitSmEncoded(sourceAddr, destinationAddr, shortMessage, params)
}

func (t *transceiverWrapper) DeliverSmResp(seq uint32, status smpp.CMDStatus) error {
	return t.tr.DeliverSmResp(seq, status)
}

func (t *transceiverWrapperFactory) GetTransceiver(host string, port int, eli int, bindParams smpp.Params) (TransceiverWrapper, error) {
	tr, err := smpp.NewTransceiver(host, port, eli, bindParams)
	if err != nil {
		return nil, err
	}
	return &transceiverWrapper{tr: tr}, nil
}

type SmppClient interface {
	Connect() error
	Disconnect()
	Reconnect() error
	IsConnected() bool
	SendMessage(id uint32, from, phone, text string) error
	BindSubmitSmResponseHandler(handler func(id, status uint32, smscId uint64))
	BindDeliverSmHandler(handler func(smscId uint64, status string))
	ReadPacket() error
}

type smppClient struct {
	smscIp           string
	smscPort         int
	smscAccount      string
	smscPassword     string
	smscEnqLnkIntrvl int
	smsMaxLen        int

	connected int32

	transceiver        TransceiverWrapper //*smpp.Transceiver
	transceiverFactory TransceiverWrapperFactory
	rateLimiter        RateLimiter
	submitSmHandler    func(id, status uint32, smscId uint64)
	deliverHandler     func(smscId uint64, status string)
}

func (c *smppClient) BindSubmitSmResponseHandler(handler func(id, status uint32, smscId uint64)) {
	c.submitSmHandler = handler
}

func (c *smppClient) BindDeliverSmHandler(handler func(smscId uint64, status string)) {
	c.deliverHandler = handler
}

func NewClient(smscIp string, smscPort int, smscAccount, smscPassword string, smscEnqLnkIntrvl, tps int) SmppClient {
	return &smppClient{
		smscIp:             smscIp,
		smscPort:           smscPort,
		smscAccount:        smscAccount,
		smscPassword:       smscPassword,
		smscEnqLnkIntrvl:   smscEnqLnkIntrvl,
		rateLimiter:        rate.NewLimiter(rate.Limit(tps), 1),
		transceiverFactory: &transceiverWrapperFactory{},
	}
}

func (c *smppClient) Disconnect() {
	defer func() {
		r := recover()
		if r != nil {
			log.Error.Println("Recovered in Disconnect", r)
		}
		atomic.StoreInt32(&c.connected, 0)
	}()

	log.Info.Println("Disconnecting from SMSC")

	if c.transceiver != nil {
		_ = c.transceiver.Unbind()
		c.transceiver.Close()
	}
}

func (c *smppClient) Connect() error {
	defer func() {
		r := recover()
		if r != nil {
			log.Error.Println("Recovered in Connect", r)
			atomic.StoreInt32(&c.connected, 0)
		}
	}()

	log.Info.Println("Connecting to SMSC")

	var err error
	c.transceiver, err = c.transceiverFactory.GetTransceiver(
		c.smscIp,
		c.smscPort,
		c.smscEnqLnkIntrvl,
		smpp.Params{
			"system_id": c.smscAccount,
			"password":  c.smscPassword,
		},
	)

	if err == nil {
		atomic.StoreInt32(&c.connected, 1)
		log.Info.Println("Connection succeeded")
	} else {
		atomic.StoreInt32(&c.connected, 0)
		log.Error.Println("Connection failed")
	}

	return err
}

func (c *smppClient) Reconnect() error {
	c.Disconnect()
	return c.Connect()
}

func (c *smppClient) IsConnected() bool {
	return atomic.LoadInt32(&c.connected) == 1
}

func (c *smppClient) SendMessage(id uint32, from, phone, text string) error {
	//impose tps limit
	c.rateLimiter.Wait(context.Background())

	defer func() {
		r := recover()
		if r != nil {
			log.Error.Println("Recovered in SendMessage", r)
			atomic.StoreInt32(&c.connected, 0)
		}
	}()

	//determine encoding
	msgEncoding := smpp.ENCODING_DEFAULT
	textBytes := []byte(text)
	partLength := 153
	maxLength := 160
	if !util.IsASCII(text) {
		msgEncoding = smpp.ENCODING_ISO10646
		textBytes = gsmutil.EncodeUcs2(text)
		partLength = 134
		maxLength = 140
	}

	textBytesLen := len(textBytes)

	if textBytesLen > maxLength {
		partsCount := int(math.Ceil(float64(textBytesLen) / float64(partLength)))

		commonId := make([]byte, 1)
		_, err := rand.Read(commonId)
		log.WarnIfErr("Error generating common id", err)

		for i := 1; i <= partsCount; i++ {
			partNo := i
			finalPart := i == partsCount
			part := []byte{0x05, 0x00, 0x03, commonId[0], byte(partsCount), byte(partNo)}
			var registeredDelivery int
			if finalPart {
				part = append(part, textBytes[(i-1)*partLength:]...)
				//set id
				c.transceiver.SetNextId(id)
				registeredDelivery = 1
			} else {
				part = append(part, textBytes[(i-1)*partLength:i*partLength]...)
				//set id
				c.transceiver.SetNextId(0)
				registeredDelivery = 0
			}

			//send
			_, err := c.transceiver.SubmitSmEncoded(from, phone, part, &smpp.Params{
				smpp.SOURCE_ADDR_TON:     5,
				smpp.SOURCE_ADDR_NPI:     1,
				smpp.DEST_ADDR_TON:       1,
				smpp.DEST_ADDR_NPI:       1,
				smpp.ESM_CLASS:           smpp.ESM_CLASS_GSMFEAT_UDHI,
				smpp.REGISTERED_DELIVERY: registeredDelivery,
				smpp.DATA_CODING:         msgEncoding,
			})

			if finalPart {
				return err
			} else {
				log.ErrIfErr("Error sending submit_sm", err)
			}
		}

	} else {
		//set id
		c.transceiver.SetNextId(id)
		//send
		_, err := c.transceiver.SubmitSmEncoded(from, phone, textBytes, &smpp.Params{
			smpp.SOURCE_ADDR_TON:     5,
			smpp.SOURCE_ADDR_NPI:     1,
			smpp.DEST_ADDR_TON:       1,
			smpp.DEST_ADDR_NPI:       1,
			smpp.REGISTERED_DELIVERY: 1,
			smpp.DATA_CODING:         msgEncoding,
		})

		return err
	}

	return nil
}

func (c *smppClient) ReadPacket() error {

	defer func() {
		r := recover()
		if r != nil {
			atomic.StoreInt32(&c.connected, 0)
			log.Error.Println("Recovered in ReadPacket", r)
		}
	}()

	pdu, err := c.transceiver.Read() // This is blocking
	if err != nil {
		if _, ok := err.(smpp.SmppErr); ok {
			log.Warn.Println("Error reading packet", err)
		} else {
			//set connected to false
			atomic.StoreInt32(&c.connected, 0)
			log.Error.Println("Error reading packet", err)
		}
		return err
	}

	// Transceiver auto handles EnquireLinks
	switch pdu.GetHeader().Id {
	case smpp.SUBMIT_SM_RESP:
		//received submit_sm_resp
		c.processSubmitSmResp(pdu)

	case smpp.DELIVER_SM:
		// received deliver_sm

		//send deliverSmResp
		err = c.transceiver.DeliverSmResp(pdu.GetHeader().Sequence, smpp.ESME_ROK)
		log.ErrIfErr("DeliverSmResp err:", err)

		c.processDeliverSm(pdu)

	default:
		log.Trace.Println("PDU ID:", pdu.GetHeader().Id)
	}

	return nil
}

func (c *smppClient) processSubmitSmResp(pdu smpp.Pdu) {
	seqId := pdu.GetHeader().Sequence
	if seqId == 0 {
		return
	}
	submitStatus := uint32(pdu.GetHeader().Status)
	msgId := pdu.GetField("message_id").String()
	smscId, err := strconv.ParseUint(msgId, 10, 64)
	if err != nil {
		smscId, err = strconv.ParseUint(msgId, 16, 64)
		if err != nil {
			log.Error.Println("Failed to parse submit_sm_rep", msgId)
			return
		}
	}

	go c.submitSmHandler(seqId, submitStatus, smscId)

	log.Info.Printf("SubmitSmResp: id=%d, smsId=%d,status=%d\n", seqId, smscId, submitStatus)
}

func (c *smppClient) processDeliverSm(pdu smpp.Pdu) {
	dlvSm := pdu.GetField("short_message").String()

	res := dlvRctRx.FindAllStringSubmatch(dlvSm, -1)
	if len(res) != 1 || len(res[0]) != 3 {
		log.Error.Println("Failed to parse deliver_sm", dlvSm)
		return
	}
	smscId, err := strconv.ParseUint(res[0][1], 10, 64)
	if err != nil {
		smscId, err = strconv.ParseUint(res[0][1], 16, 64)
		if err != nil {
			log.Error.Println("Failed to parse deliver_sm", dlvSm)
			return
		}
	}
	status := res[0][2]

	go c.deliverHandler(smscId, status)

	log.Info.Printf("DeliverSm: smscId=%d, status=%s\n", smscId, status)
}
