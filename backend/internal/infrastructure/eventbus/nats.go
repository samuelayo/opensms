package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/samuelayo/opensms/internal/infrastructure/config"
)

// EventBus defines the interface for event publishing and subscribing
type EventBus interface {
	Publish(ctx context.Context, subject string, data interface{}) error
	Subscribe(subject string, handler MessageHandler) (*Subscription, error)
	QueueSubscribe(subject, queue string, handler MessageHandler) (*Subscription, error)
	Close()
}

// MessageHandler is a function that handles received messages
type MessageHandler func(ctx context.Context, msg *Message) error

// Message represents an event message
type Message struct {
	ID        string                 `json:"id"`
	Subject   string                 `json:"subject"`
	Data      json.RawMessage        `json:"data"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
}

// Subscription represents an event subscription
type Subscription struct {
	sub *nats.Subscription
}

// Unsubscribe unsubscribes from the event
func (s *Subscription) Unsubscribe() error {
	return s.sub.Unsubscribe()
}

// NATSConnection wraps nats.Conn with additional functionality
type NATSConnection struct {
	conn *nats.Conn
}

// NewNATSConnection creates a new NATS connection
func NewNATSConnection(cfg config.NATSConfig) (*NATSConnection, error) {
	opts := []nats.Option{
		nats.MaxReconnects(cfg.MaxReconnects),
		nats.ReconnectWait(cfg.ReconnectWait),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				fmt.Printf("NATS disconnected: %v\n", err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			fmt.Printf("NATS reconnected to %s\n", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			fmt.Println("NATS connection closed")
		}),
	}

	conn, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return &NATSConnection{conn: conn}, nil
}

// Publish publishes an event to a subject
func (n *NATSConnection) Publish(ctx context.Context, subject string, data interface{}) error {
	msg := &Message{
		ID:        generateID(),
		Subject:   subject,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"publisher": "opensms",
		},
	}

	// Marshal data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}
	msg.Data = jsonData

	// Marshal message
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Publish with timeout
	if err := n.conn.Publish(subject, msgBytes); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

// Subscribe subscribes to events on a subject
func (n *NATSConnection) Subscribe(subject string, handler MessageHandler) (*Subscription, error) {
	sub, err := n.conn.Subscribe(subject, func(m *nats.Msg) {
		var msg Message
		if err := json.Unmarshal(m.Data, &msg); err != nil {
			fmt.Printf("Failed to unmarshal message: %v\n", err)
			return
		}

		ctx := context.Background()
		if err := handler(ctx, &msg); err != nil {
			fmt.Printf("Handler error for subject %s: %v\n", subject, err)
		}
	})

	if err != nil {
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}

	return &Subscription{sub: sub}, nil
}

// QueueSubscribe subscribes to events with queue semantics (load balancing)
func (n *NATSConnection) QueueSubscribe(subject, queue string, handler MessageHandler) (*Subscription, error) {
	sub, err := n.conn.QueueSubscribe(subject, queue, func(m *nats.Msg) {
		var msg Message
		if err := json.Unmarshal(m.Data, &msg); err != nil {
			fmt.Printf("Failed to unmarshal message: %v\n", err)
			return
		}

		ctx := context.Background()
		if err := handler(ctx, &msg); err != nil {
			fmt.Printf("Handler error for subject %s: %v\n", subject, err)
		}
	})

	if err != nil {
		return nil, fmt.Errorf("failed to queue subscribe: %w", err)
	}

	return &Subscription{sub: sub}, nil
}

// Close closes the NATS connection
func (n *NATSConnection) Close() {
	if n.conn != nil {
		n.conn.Close()
	}
}

// Health checks NATS health
func (n *NATSConnection) Health() error {
	if n.conn.Status() != nats.CONNECTED {
		return fmt.Errorf("NATS not connected, status: %v", n.conn.Status())
	}
	return nil
}

// Event subjects (define your event types here)
const (
	// User events
	SubjectUserCreated   = "user.created"
	SubjectUserUpdated   = "user.updated"
	SubjectUserDeleted   = "user.deleted"
	SubjectUserLoggedIn  = "user.logged_in"
	SubjectUserLoggedOut = "user.logged_out"

	// Student events
	SubjectStudentEnrolled  = "student.enrolled"
	SubjectStudentWithdrawn = "student.withdrawn"
	SubjectStudentPromoted  = "student.promoted"

	// Academic events
	SubjectGradePublished    = "grade.published"
	SubjectAttendanceMarked  = "attendance.marked"
	SubjectAssignmentCreated = "assignment.created"
	SubjectExamScheduled     = "exam.scheduled"

	// Finance events
	SubjectFeeGenerated = "fee.generated"
	SubjectPaymentMade  = "payment.made"
	SubjectInvoiceSent  = "invoice.sent"

	// Notification events
	SubjectNotificationSend = "notification.send"
	SubjectEmailSend        = "email.send"
	SubjectSMSSend          = "sms.send"
)

// Helper function to generate unique IDs
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
