package email

import "log"

type Service interface {
	SendConfirmationEmail(toEmail, city, confirmationLink string) error
	SendWeatherUpdateEmail(toEmail, city, weatherInfo, unsubscribeLink string) error
}

// for now just a dummy email service that logs to console.
type LogEmailService struct{}

func NewLogEmailService() *LogEmailService {
	return &LogEmailService{}
}

// TODO: change these send actual e-mails later
func (s *LogEmailService) SendConfirmationEmail(toEmail, city, confirmationLink string) error {
	log.Printf("--- SENDING CONFIRMATION EMAIL ---")
	log.Printf("To: %s", toEmail)
	log.Printf("City: %s", city)
	log.Printf("Subject: Confirm your Weather Subscription for %s", city)
	log.Printf("Body: Please confirm your subscription by clicking this link: %s", confirmationLink)
	log.Printf("--- END EMAIL ---")
	return nil
}

func (s *LogEmailService) SendWeatherUpdateEmail(toEmail, city, weatherInfo, unsubscribeLink string) error {
	log.Printf("--- SENDING WEATHER UPDATE EMAIL ---")
	log.Printf("To: %s", toEmail)
	log.Printf("City: %s", city)
	log.Printf("Subject: Your Weather Update for %s", city)
	log.Printf("Body: %s\n\nUnsubscribe: %s", weatherInfo, unsubscribeLink)
	log.Printf("--- END EMAIL ---")
	return nil
}
