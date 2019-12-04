[![Build Status](https://img.shields.io/travis/Dilshat/sms-sender/master.svg)](https://travis-ci.org/Dilshat/sms-sender)

#### SMS service
Simple sms service that provides HTTP API to send text messages.
It connects directly to SMS Center or SMS proxy of telecom operator using SMPP 3.4 procotol.

HTTP API description is available at `http://${base-path}/swagger/index.html` (the service must be running)

#### Examples of using HTTP API

- Sending message (there might be more than one recipient phone):
```
curl localhost:8080/sms -H "Content-Type: application/json" -d '{"phones":["996778295555"],"text":"hello", "sender":"awesome"}'
```
will return response containing id of message, which can be used later to check status of message delivery:
```
{"id":56}
```

- Check status of message delivery
```
curl localhost:8080/sms/56
curl localhost:8080/sms/56?phone=996778295555
```
response:
```
{
  "id": 56,
  "sender": "awesome",
  "text": "hello",
  "statuses": [
    {
      "phone": "996778295555",
      "status": "DELIVRD"
    }
  ]
}
```

Message statues are stored N days in the service database (_number of days can be configured in the service settings_).

All settings are stored in the file **.env**; environment variables with the same names as in the .env file override the latter ones.

#### Delivery status reception

If _WEB_HOOK_ is set to some non-empty URL, the service will send notifications about delivery status receipt (a separate update per each phone) to the specified http endpoint in the following form:

```
{
  "id": 56,
  "sender": "awesome",
  "text": "hello",
  "statuses": [
    {
      "phone": "996778295555",
      "status": "DELIVRD"
    }
  ]
}
```