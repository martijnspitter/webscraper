package reporting

import (
	"fmt"
	"huurwoning/config"
	"huurwoning/logger"
	"net/smtp"

	"github.com/jordan-wright/email"
	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

func SendAlert(newAdress string, prefix string, logger *logger.Logger) {
	body := prefix + " New adress found: " + newAdress
	res, err := sendSMS(body)
	if err != nil {
		logger.Error("Error sending SMS", "error", err)
	} else {
		logger.Info("SMS sent", "response", res)
	}

	res, err = sendEmail(body, body)
	if err != nil {
		logger.Error("Error sending email", "error", err)
	} else {
		logger.Info("Email sent", "response", res)
	}
}

func SendAlertForMultipleResults(results string, prefix string, logger *logger.Logger) {
	body := prefix + " New adress found: \n" + results
	subject := prefix + " Multiple new results found!"
	res, err := sendEmail(body, subject)
	if err != nil {
		logger.Error("Error sending email", "error", err)
	} else {
		logger.Info("Email sent", "response", res)
	}
}

func sendSMS(body string) (string, error) {
	config, err := config.LoadConfig()
	if err != nil {
		return "", fmt.Errorf("Failed to load config: %v", err)
	}
	// send an sms with twilio.
	// add the prefix to the newAdress
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: config.TWILIO_SID,
		Password: config.TWILIO_TOKEN,
	})

	params := &openapi.CreateMessageParams{}
	params.SetTo(config.YOUR_PHONE_NUMBER)
	params.SetFrom(config.TWILIO_PHONE_NUMBER)
	params.SetBody(body)

	resp, err := client.Api.CreateMessage(params)
	if err != nil {
		return "", fmt.Errorf("Error sending SMS: %v", err)
	}

	return *resp.Sid, nil
}

func sendEmail(body string, subject string) (string, error) {
	config, err := config.LoadConfig()
	if err != nil {
		return "", fmt.Errorf("Failed to load config: %v", err)
	}

	e := email.NewEmail()
	e.From = config.FROM_EMAIL
	e.To = []string{config.TO_EMAIL}
	e.Subject = subject
	e.Text = []byte(body)

	// Set headers to mark the email as important
	e.Headers.Add("X-Priority", "1")    // 1 = High, 3 = Normal, 5 = Low
	e.Headers.Add("Importance", "High") // High, Normal, Low

	err = e.Send(fmt.Sprintf("%s:%d", config.SMTP_SERVER, config.SMTP_PORT), smtp.PlainAuth("", config.SMTP_USERNAME, config.SMTP_PASSWORD, config.SMTP_SERVER))
	if err != nil {
		return "", fmt.Errorf("Error sending email: %v", err)
	} else {
		return fmt.Sprintf("Email sent successfully"), nil
	}
}
