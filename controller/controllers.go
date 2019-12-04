package controller

import (
	"github.com/dilshat/sms-sender/log"
	"github.com/dilshat/sms-sender/service"
	"github.com/dilshat/sms-sender/service/dto"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"strings"
)

// SendSms godoc
// @Summary Send sms
// @Description Sends sms message to specified phones
// @Accept json
// @Produce json
// @Param sms body dto.Message true "Message"
// @Success 200 {object} dto.Id
// @Failure 400 "error description"
// @Router /sms [post]
func GetSendSmsFunc(srv service.Service) echo.HandlerFunc {

	return func(c echo.Context) error {
		msg := new(dto.Message)
		if err := c.Bind(msg); err != nil {
			return err
		}

		id, err := srv.SendMessage(*msg)
		if err != nil {
			switch err.(type) {
			case *service.InvalidPayloadErr:
				return c.String(http.StatusBadRequest, err.Error())
			default:
				log.Error.Println(err)
				return c.String(http.StatusInternalServerError, "System malfunction. Please, try later")
			}
		}

		return c.JSON(http.StatusOK, id)
	}
}

// CheckSms godoc
// @Summary Check sms
// @Description Checks sms message delivery status
// @Produce json
// @Param id path int true "Message id"
// @Param phone query string false "Phone number"
// @Success 200 {object} dto.MessageStatus
// @Failure 400 "error description"
// @Router /sms/{id} [get]
func GetCheckSmsFunc(service service.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		phone := c.QueryParam("phone")

		id64, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return err
		}
		id32 := uint32(id64)

		if strings.TrimSpace(phone) == "" {
			status, err := service.CheckStatusOfMessage(id32)
			if err != nil {
				if err.Error() == "not found" {
					return c.String(http.StatusNotFound, "Message not found "+id)
				} else {
					log.Error.Println(err)
					return c.String(http.StatusInternalServerError, "System malfunction. Please, try later")
				}
			}

			return c.JSON(http.StatusOK, status)
		} else {
			status, err := service.CheckStatusOfRecipient(id32, phone)
			if err != nil {
				if err.Error() == "not found" {
					return c.String(http.StatusNotFound, "Phone not found "+phone)
				} else {
					log.Error.Println(err)
					return c.String(http.StatusInternalServerError, "System malfunction. Please, try later")
				}
			}

			return c.JSON(http.StatusOK, status)
		}

	}
}
