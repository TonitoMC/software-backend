package scheduler

import (
	"context"
	"log"
	"time"

	"software-backend/internal/service"
)

// ReminderScheduler runs periodic checks for sending WhatsApp reminders
type ReminderScheduler struct {
	whatsAppService service.WhatsAppService
	interval        time.Duration
	stopChan        chan struct{}
}

// NewReminderScheduler creates a new reminder scheduler
func NewReminderScheduler(whatsAppService service.WhatsAppService, interval time.Duration) *ReminderScheduler {
	return &ReminderScheduler{
		whatsAppService: whatsAppService,
		interval:        interval,
		stopChan:        make(chan struct{}),
	}
}

// Start begins the scheduler
func (s *ReminderScheduler) Start() {
	log.Println("Starting WhatsApp reminder scheduler...")
	ticker := time.NewTicker(s.interval)

	// Run immediately on start
	s.runCheck()

	go func() {
		for {
			select {
			case <-ticker.C:
				s.runCheck()
			case <-s.stopChan:
				ticker.Stop()
				log.Println("WhatsApp reminder scheduler stopped")
				return
			}
		}
	}()
}

// Stop stops the scheduler
func (s *ReminderScheduler) Stop() {
	close(s.stopChan)
}

func (s *ReminderScheduler) runCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	log.Println("Checking for upcoming appointments...")
	if err := s.whatsAppService.CheckAndScheduleReminders(ctx); err != nil {
		log.Printf("Error checking reminders: %v", err)
	}

	// Also retry any pending notifications
	if err := s.whatsAppService.ProcessPendingReminders(ctx); err != nil {
		log.Printf("Error processing pending reminders: %v", err)
	}
}
