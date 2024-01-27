package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/mail"
	"os"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/joho/godotenv"
)

var (
	username = flag.String("u", "", "Username for gmail account")
	password = flag.String("p", "", "Password (or app password) for gmail account")
	destdir  = flag.String("d", "", "Destination directory for IMAP structure")
)

func main() {
	flag.Parse()

	if *destdir == "" {
		flag.PrintDefaults()
		return
	}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	un := *username
	if un == "" {
		un = os.Getenv("USERNAME")
	}
	pw := *password
	if pw == "" {
		pw = os.Getenv("USERNAME")
	}

	log.Println("Connecting to server...")

	// Connect to server
	c, err := client.DialTLS("imap.gmail.com:993", nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	if err := c.Login(un, pw); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	defer c.Logout()

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	mboxes := []*imap.MailboxInfo{}
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	log.Println("Mailboxes:")
	for m := range mailboxes {
		mboxes = append(mboxes, m)
		log.Printf("* %s", m.Name)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	// ------

	for _, m := range mboxes {
		// Select INBOX
		mbox, err := c.Select(strings.TrimSpace(m.Name), false)
		if err != nil {
			log.Printf("ERR: %s", err.Error())
			continue
		}
		log.Printf("Opened '%s' for reading [%d messages]", m.Name, mbox.Messages)
		log.Printf("Flags for %s: %v", m.Name, mbox.Flags)

		if mbox.Messages == 0 {
			log.Printf("No messages, skipping mailbox")
			continue
		}

		// Get the last 4 messages
		from := uint32(1)
		to := mbox.Messages
		if mbox.Messages > 3 {
			// We're using unsigned integers here, only subtract if the result is > 0
			from = mbox.Messages - 3
		}
		seqset := new(imap.SeqSet)
		seqset.AddRange(from, to)

		messages := make(chan *imap.Message, 10)
		done = make(chan error, 1)
		go func() {
			done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
		}()

		log.Println("Last 4 messages:")
		for msg := range messages {
			log.Println("* " + msg.Envelope.Subject)
		}

		if err := <-done; err != nil {
			log.Printf("ERR: %s", err.Error())
		}

		fetch(c, mbox)
	}
}

func fetch(c *client.Client, mbox *imap.MailboxStatus) error {
	// Get the last message
	if mbox.Messages == 0 {
		log.Printf("%s: No messages in mailbox", mbox.Name)
		return fmt.Errorf("no messages in mailbox %s", mbox.Name)
	}
	seqset := new(imap.SeqSet)
	seqset.AddRange(mbox.Messages, mbox.Messages)

	// Get the whole message body
	section := &imap.BodySectionName{}
	items := []imap.FetchItem{section.FetchItem()}

	messages := make(chan *imap.Message, 1)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, items, messages)
	}()

	log.Println("Last message:")
	msg := <-messages
	r := msg.GetBody(section)
	if r == nil {
		err := fmt.Errorf("server didn't return message body")
		log.Printf("%s: ERR: %s", mbox.Name, err.Error())
		return err
	}

	if err := <-done; err != nil {
		log.Printf("%s: ERR: %s", mbox.Name, err.Error())
		return err
	}

	m, err := mail.ReadMessage(r)
	if err != nil {
		log.Printf("%s: ERR: %s", mbox.Name, err.Error())
		return err
	}

	/*
		header := m.Header
		log.Println("Date:", header.Get("Date"))
		log.Println("From:", header.Get("From"))
		log.Println("To:", header.Get("To"))
		log.Println("Subject:", header.Get("Subject"))
	*/
	log.Printf("%#v", m.Header)

	id := m.Header.Get("Message-Id")
	log.Printf("MESSAGE : %s", id)
	body, err := io.ReadAll(m.Body)
	if err != nil {
		log.Printf("%s: ERR: %s", mbox.Name, err.Error())
		return err
	}
	log.Printf("Message body length %d", len(string(body)))
	return nil
}
