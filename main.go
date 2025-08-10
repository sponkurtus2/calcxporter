package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/gocarina/gocsv"
	"github.com/joho/godotenv"
	"github.com/resend/resend-go/v2"
)

type Urgency string

const (
	Normal    Urgency = "normal"
	Important Urgency = "important"
)

// Structure of the event
// 0,2025,8,1,"Going for a matcha 🍵",1,once,normal
type Event struct {
	Id            int
	Year          int
	Month         int
	Day           int
	EventName     string
	NotUsedInt    string `csv:"-"`
	NotUsedString string `csv:"-"`
	Urgency       Urgency
}

type ByDate []*Event

func (a ByDate) Len() int      { return len(a) }
func (a ByDate) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByDate) Less(i, j int) bool {
	if a[i].Year != a[j].Year {
		return a[i].Year > a[j].Year
	}
	if a[i].Month != a[j].Month {
		return a[i].Month > a[j].Month
	}
	return a[i].Day > a[j].Day
}

func createFile() {
	csvHeaders := "Id,Year,Month,Day,EventName,NotUsedInt,NotUsedString,Urgency\n"
	csvHeadersBytes := []byte(csvHeaders)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	fileLocation := fmt.Sprintf("%s/.config/calcure/events.csv", homeDir)
	newFileLocation := "./events.csv"

	bytesRead, err := os.ReadFile(fileLocation)
	if err != nil {
		panic(err)
	}

	fullContent := append(csvHeadersBytes, bytesRead...)
	err = os.WriteFile(newFileLocation, fullContent, 0644)
	if err != nil {
		panic(err)
	}
}

func readFile() []*Event {
	csvFile, err := os.OpenFile("events.csv", os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()

	var events []*Event

	if err := gocsv.UnmarshalFile(csvFile, &events); err != nil {
		panic(err)
	}

	return events
}

func createHtml(events []*Event) string {
	sort.Sort(ByDate(events))
	htmlContent := `<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Eingehende Ereignisse</title>
    <style>
        body {
            font-family: 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 650px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f9f9f9;
        }
        .header {
            text-align: center;
            margin-bottom: 30px;
            padding-bottom: 20px;
            border-bottom: 1px solid #eaeaea;
        }
        .header h1 {
            font-weight: 300;
            color: #2c3e50;
            margin-bottom: 5px;
        }
        .header p {
            color: #7f8c8d;
            margin-top: 0;
        }
        .event {
            background: white;
            padding: 20px;
            margin-bottom: 15px;
            border-radius: 4px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.05);
            transition: all 0.3s ease;
        }
        .event:hover {
            box-shadow: 0 5px 15px rgba(0,0,0,0.1);
        }
        .event-date {
            font-size: 14px;
            color: #7f8c8d;
            margin-bottom: 5px;
        }
        .event-name {
            font-size: 18px;
            font-weight: 500;
            color: #2c3e50;
            margin: 0 0 10px 0;
        }
        .event-urgency {
            display: inline-block;
            padding: 3px 8px;
            font-size: 12px;
            border-radius: 3px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        .normal {
            background-color: #ecf0f1;
            color: #7f8c8d;
        }
        .important {
            background-color: #ffeaa7;
            color: #e17055;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Kommende Veranstaltungen</h1>
        <p>Dies sind Ihre wichtigsten Veranstaltungen für heute</p>
    </div>`

	for _, event := range events {
		urgencyClass := "normal"
		urgencyText := "Normal"
		if event.Urgency == "important" {
			urgencyClass = "important"
			urgencyText = "Wichtig"
		}

		htmlContent += fmt.Sprintf(`
    <div class="event">
        <div class="event-date">%d/%d/%d</div>
        <h2 class="event-name">%s</h2>
        <span class="event-urgency %s">%s</span>
    </div>`, event.Day, event.Month, event.Year, event.EventName, urgencyClass, urgencyText)
	}

	// Close the HTML tags
	htmlContent += `
</body>
</html>`

	return htmlContent
}

func sendEmail(eventsHtml string) {
	API_KEY := os.Getenv("API_KEY")
	EMAIL := os.Getenv("EMAIL")

	client := resend.NewClient(API_KEY)

	emailParams := &resend.SendEmailRequest{
		To:      []string{EMAIL},
		From:    "remainders@carlinux.me",
		Html:    eventsHtml,
		Subject: "Daily reminder :)",
	}

	sent, err := client.Emails.Send(emailParams)
	if err != nil {
		panic(err)
	}
	fmt.Println(sent.Id)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	createFile()

	events := readFile()

	text := createHtml(events)
	sendEmail(text)
}
