package controller

import (
	"errors"
	"github.com/dilshat/sms-sender/service"
	"github.com/dilshat/sms-sender/service/dto"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"testing"
)

var (
	OK200        bool
	stringCalled bool
)

func TestGetSendSmsFunc(t *testing.T) {
	OK200 = false
	f := GetSendSmsFunc(mockService{})

	err := f(mockContext{})

	require.NoError(t, err)
	require.True(t, OK200)

	bindError := errors.New("Bind error")

	err = f(mockContext{bindError: bindError})

	require.Equal(t, bindError, err)

	stringCalled = false
	f = GetSendSmsFunc(&mockService{sendMsgErr: service.NewInvalidPayloadError("blablabla")})

	err = f(mockContext{})

	require.True(t, stringCalled)

	stringCalled = false
	f = GetSendSmsFunc(&mockService{sendMsgErr: errors.New("blablabla")})

	err = f(mockContext{})

	require.True(t, stringCalled)
}

func TestGetCheckSmsFunc(t *testing.T) {
	OK200 = false
	f := GetCheckSmsFunc(mockService{})

	err := f(mockContext{param: "123"})

	require.NoError(t, err)
	require.True(t, OK200)

	err = f(mockContext{param: ""})

	require.Error(t, err)

	stringCalled = false
	f = GetCheckSmsFunc(mockService{checkStatusErr: errors.New("not found")})

	err = f(mockContext{param: "123"})

	require.True(t, stringCalled)

	stringCalled = false
	f = GetCheckSmsFunc(mockService{checkStatusErr: errors.New("blablabla")})

	err = f(mockContext{param: "123"})

	require.True(t, stringCalled)

	stringCalled = false
	f = GetCheckSmsFunc(mockService{checkStatusErr: errors.New("not found")})

	err = f(mockContext{param: "123", queryParam: "996777123456"})

	require.True(t, stringCalled)

	stringCalled = false
	f = GetCheckSmsFunc(mockService{checkStatusErr: errors.New("blablabla")})

	err = f(mockContext{param: "123", queryParam: "996777123456"})

	require.True(t, stringCalled)

	OK200 = false
	f = GetCheckSmsFunc(mockService{})

	err = f(mockContext{param: "123", queryParam: "996777123456"})

	require.True(t, OK200)
}

//-----------mocks--------
type mockContext struct {
	bindError  error
	param      string
	queryParam string
}

type mockService struct {
	sendMsgErr     error
	checkStatusErr error
}

func (m mockService) SendMessage(message dto.Message) (dto.Id, error) {
	return dto.Id{}, m.sendMsgErr
}

func (m mockService) CheckStatusOfMessage(id uint32) (dto.MessageStatus, error) {
	return dto.MessageStatus{}, m.checkStatusErr
}

func (m mockService) CheckStatusOfRecipient(id uint32, phone string) (dto.MessageStatus, error) {
	return dto.MessageStatus{}, m.checkStatusErr
}

func (m mockContext) Request() *http.Request {
	panic("implement me")
}

func (m mockContext) SetRequest(r *http.Request) {
	panic("implement me")
}

func (m mockContext) SetResponse(r *echo.Response) {
	panic("implement me")
}

func (m mockContext) Response() *echo.Response {
	panic("implement me")
}

func (m mockContext) IsTLS() bool {
	panic("implement me")
}

func (m mockContext) IsWebSocket() bool {
	panic("implement me")
}

func (m mockContext) Scheme() string {
	panic("implement me")
}

func (m mockContext) RealIP() string {
	panic("implement me")
}

func (m mockContext) Path() string {
	panic("implement me")
}

func (m mockContext) SetPath(p string) {
	panic("implement me")
}

func (m mockContext) Param(name string) string {
	return m.param
}

func (m mockContext) ParamNames() []string {
	panic("implement me")
}

func (m mockContext) SetParamNames(names ...string) {
	panic("implement me")
}

func (m mockContext) ParamValues() []string {
	panic("implement me")
}

func (m mockContext) SetParamValues(values ...string) {
	panic("implement me")
}

func (m mockContext) QueryParam(name string) string {
	return m.queryParam
}

func (m mockContext) QueryParams() url.Values {
	panic("implement me")
}

func (m mockContext) QueryString() string {
	panic("implement me")
}

func (m mockContext) FormValue(name string) string {
	panic("implement me")
}

func (m mockContext) FormParams() (url.Values, error) {
	panic("implement me")
}

func (m mockContext) FormFile(name string) (*multipart.FileHeader, error) {
	panic("implement me")
}

func (m mockContext) MultipartForm() (*multipart.Form, error) {
	panic("implement me")
}

func (m mockContext) Cookie(name string) (*http.Cookie, error) {
	panic("implement me")
}

func (m mockContext) SetCookie(cookie *http.Cookie) {
	panic("implement me")
}

func (m mockContext) Cookies() []*http.Cookie {
	panic("implement me")
}

func (m mockContext) Get(key string) interface{} {
	panic("implement me")
}

func (m mockContext) Set(key string, val interface{}) {
	panic("implement me")
}

func (m mockContext) Bind(i interface{}) error {
	return m.bindError
}

func (m mockContext) Validate(i interface{}) error {
	panic("implement me")
}

func (m mockContext) Render(code int, name string, data interface{}) error {
	panic("implement me")
}

func (m mockContext) HTML(code int, html string) error {
	panic("implement me")
}

func (m mockContext) HTMLBlob(code int, b []byte) error {
	panic("implement me")
}

func (m mockContext) String(code int, s string) error {
	stringCalled = true
	return nil
}

func (m mockContext) JSON(code int, i interface{}) error {
	OK200 = true
	return nil
}

func (m mockContext) JSONPretty(code int, i interface{}, indent string) error {
	panic("implement me")
}

func (m mockContext) JSONBlob(code int, b []byte) error {
	panic("implement me")
}

func (m mockContext) JSONP(code int, callback string, i interface{}) error {
	panic("implement me")
}

func (m mockContext) JSONPBlob(code int, callback string, b []byte) error {
	panic("implement me")
}

func (m mockContext) XML(code int, i interface{}) error {
	panic("implement me")
}

func (m mockContext) XMLPretty(code int, i interface{}, indent string) error {
	panic("implement me")
}

func (m mockContext) XMLBlob(code int, b []byte) error {
	panic("implement me")
}

func (m mockContext) Blob(code int, contentType string, b []byte) error {
	panic("implement me")
}

func (m mockContext) Stream(code int, contentType string, r io.Reader) error {
	panic("implement me")
}

func (m mockContext) File(file string) error {
	panic("implement me")
}

func (m mockContext) Attachment(file string, name string) error {
	panic("implement me")
}

func (m mockContext) Inline(file string, name string) error {
	panic("implement me")
}

func (m mockContext) NoContent(code int) error {
	panic("implement me")
}

func (m mockContext) Redirect(code int, url string) error {
	panic("implement me")
}

func (m mockContext) Error(err error) {
	panic("implement me")
}

func (m mockContext) Handler() echo.HandlerFunc {
	panic("implement me")
}

func (m mockContext) SetHandler(h echo.HandlerFunc) {
	panic("implement me")
}

func (m mockContext) Logger() echo.Logger {
	panic("implement me")
}

func (m mockContext) Echo() *echo.Echo {
	panic("implement me")
}

func (m mockContext) Reset(r *http.Request, w http.ResponseWriter) {
	panic("implement me")
}
