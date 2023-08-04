package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func googleCalendar() {
	CalendarId := os.Getenv("GOOGLE_MED_CALENDAR")
	jsonKey, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("unable to read json key file: %v", err)
	}

	creds, err := google.CredentialsFromJSON(context.Background(), jsonKey, calendar.CalendarScope)
	if err != nil {
		log.Fatalf("unable to parse client secret file to config: %v", err)
	}

	ctx := context.Background()
	calendarService, err := calendar.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	t := time.Now().Format(time.RFC3339)
	events, err := calendarService.Events.List(CalendarId).ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		log.Printf("Unable to retrieve next ten of the user's events. %v: %v", err, events)
	}

	fmt.Println("Upcoming events:")
	if events == nil || len(events.Items) == 0 {
		fmt.Println("No upcoming events found.")
	} else {
		for _, item := range events.Items {
			date := item.Start.DateTime
			if date == "" {
				date = item.Start.Date
			}
			fmt.Printf("%v (%v)\n", item.Summary, date)
		}
	}

	event := &calendar.Event{
		Summary:  "Google I/O 2025",
		Location: "800 Howard St., San Francisco, CA 94103",
		Start: &calendar.EventDateTime{
			DateTime: time.Now().Add(time.Hour * 24).Format(time.RFC3339),
			TimeZone: "America/Los_Angeles",
		},
		End: &calendar.EventDateTime{
			DateTime: time.Now().Add(time.Hour * 26).Format(time.RFC3339),
			TimeZone: "America/Los_Angeles",
		},
	}

	newEvent, err := calendarService.Events.Insert(CalendarId, event).Do()
	if err != nil {
		log.Fatalf("Unable to create event: %v: %v", err, newEvent)
	}
	fmt.Printf("Event created: %s\n", newEvent.HtmlLink)

}
