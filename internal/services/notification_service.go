package services

import (
	"log"
)

// In a real application, this would integrate with WhatsApp APIs like Fonnte or Twilio.
// For now, we simulate async processing with goroutines.

func NotifyNewReport(reportID uint) {
	go func() {
		log.Printf("[ASYNC NOTIFY] New SOS Report received (ID: %d). Alerting relawan...", reportID)
		// TODO: fetch all relawan phones and send WA messages
	}()
}

func NotifyReportStatusChange(reportID uint, status string) {
	go func() {
		log.Printf("[ASYNC NOTIFY] Report ID %d status changed to: %s. Alerting korban...", reportID, status)
		// TODO: fetch korban phone and send WA message
	}()
}

func NotifyDistribution(reportID uint, item string, qty int) {
	go func() {
		log.Printf("[ASYNC NOTIFY] Bantuan dikirim (Report ID %d): %d %s. Alerting korban...", reportID, qty, item)
	}()
}
