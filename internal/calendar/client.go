package calendar

import (
	"google.golang.org/api/calendar/v3"
)

// Client wraps the Google Calendar service
type Client struct {
	service *calendar.Service
}

// NewClient creates a new calendar client
func NewClient(service *calendar.Service) *Client {
	return &Client{service: service}
}

// GetService returns the underlying calendar service
func (c *Client) GetService() *calendar.Service {
	return c.service
}
