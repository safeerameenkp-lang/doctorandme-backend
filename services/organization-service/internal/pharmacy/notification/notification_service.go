package notification

type NotificationService interface {
	SendSMS(phone, message string) error
	SendPush(token, title, body string) error
}

type notificationService struct {
	// Repo dependency
}

func NewNotificationService() NotificationService {
	return &notificationService{}
}

func (s *notificationService) SendSMS(phone, message string) error {
	return nil
}

func (s *notificationService) SendPush(token, title, body string) error {
	return nil
}
