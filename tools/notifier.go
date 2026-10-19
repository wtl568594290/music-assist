package tools

import "github.com/wailsapp/wails/v3/pkg/services/notifications"

const (
	maxRequestCnt = 3
)

var requestCnt int = 0

func SendNotification(notifier *notifications.NotificationService, id string, title, body string) {
	if notifier == nil {
		return
	}
	if authorized, err := notifier.CheckNotificationAuthorization(); err == nil {
		if !authorized {
			if requestCnt >= maxRequestCnt {
				return
			}
			// Request authorization
			authorized, err = notifier.RequestNotificationAuthorization()
			requestCnt++
		}
		if authorized {
			// Send notifications
			notifier.SendNotification(notifications.NotificationOptions{
				ID:    id,
				Title: title,
				Body:  body,
			})
		}
	}
}
