package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

// ============================================================
// CONFIGURATION
// ============================================================

const (
	// IRC server details
	SERVER = "chat.freenode.net:6667"
	//Your bot details
	NICKNAME = "CCClient123"
	USERNAME = "guest"
	REALNAME = "Coding Challenges Client"
)

// ============================================================
// DATA STRUCTURES
// ============================================================

// Message represents a parsed IRC message
type Message struct {
	Prefix  string   //sender
	Command string   //type of message(PING, 001, PRIVMSG, etc..)
	Params  []string //Message parameters (Additional data)
}

// IRCClient holds the client state
type IRCClient struct {
	conn           net.Conn
	currentChannel string
	nickname       string
}

// ============================================================
// MAIN FUNCTION
// ============================================================

func main() {

	fmt.Println("Connecting to IRC server:", SERVER)

	//connect to IRC server via TCP
	conn, err := net.Dial("tcp", SERVER)
	if err != nil {
		log.Fatal("Error connecting to server:", err)
	}
	defer conn.Close()

	fmt.Println("Connected to IRC server!")

	//create client state
	client := &IRCClient{
		conn:     conn,
		nickname: NICKNAME,
	}

	//send NICK command
	sendCommand(conn, fmt.Sprintf("NICK %s", NICKNAME))
	fmt.Printf("Sent: NICK %s\n", NICKNAME)

	//send USER command
	//Format: USER <username> <mode> <unused> :<realname>
	sendCommand(conn, fmt.Sprintf("USER %s 0 * :%s", USERNAME, REALNAME))
	fmt.Printf("Sent: USER %s 0 * :%s\n\n", USERNAME, REALNAME)

	//create channel for communication between goroutines
	done := make(chan bool)

	//start goroutine to handle server messages
	go handleServerMessages(client, done)

	//Main goroutine handles user input
	handleUserInput(client, done)
}

// ============================================================
// HIGH-LEVEL ORCHESTRATION
// ============================================================

// reads commands from user
func handleUserInput(client *IRCClient, done chan bool) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n=== IRC Client Ready ===")
	fmt.Println("Commands:")
	fmt.Println("  /join #channel  - Join a channel")
	fmt.Println("  /part           - Leave current channel")
	fmt.Println("  /quit           - Disconnect from IRC")
	fmt.Println()
	fmt.Println("Type any message (without /) to send to current channel")
	fmt.Println()

	for {
		//check if server disconnected
		select {
		case <-done:
			fmt.Println("\nDisconnected from server")
			return
		default:
			//continue to read input
		}

		// Show prompt with current nickname
		if client.currentChannel != "" {
			fmt.Printf("[%s@%s] > ", client.nickname, client.currentChannel)
		} else {
			fmt.Printf("[%s] > ", client.nickname)
		}

		//read user input
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading input:", err)
			return
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		//Process command or message
		if strings.HasPrefix(input, "/") {
			handleUserCommand(client, input)
		} else {
			// User typed a regular message
			sendChatMessage(client, input)
		}
	}
}

// reads and processes messages from server
func handleServerMessages(client *IRCClient, done chan bool) {

	reader := bufio.NewReader(client.conn)

	for {
		rawMessage, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Connection closed:", err)
			done <- true
			return
		}

		//Trim whitespace
		rawMessage = strings.TrimSpace(rawMessage)
		if rawMessage == "" {
			continue
		}

		//Parse the message
		msg := parseMessage(rawMessage)

		//Handle different message types
		handleMessage(client, msg, rawMessage)
	}
}

// ============================================================
// MESSAGE SENDING
// ============================================================

// sends a message to the current channel
func sendChatMessage(client *IRCClient, message string) {
	// Check if user is in a channel
	if client.currentChannel == "" {
		fmt.Println("❌ You must join a channel first. Use /join #channel")
		return
	}

	// Validate message is not empty
	if strings.TrimSpace(message) == "" {
		return
	}

	// Format: PRIVMSG <channel> :<message>
	ircMessage := fmt.Sprintf("PRIVMSG %s :%s", client.currentChannel, message)

	// Send to server
	sendCommand(client.conn, ircMessage)

	// Display locally (instant feedback)
	fmt.Printf("[%s] %s: %s\n", client.currentChannel, client.nickname, message)
}

// ============================================================
// COMMAND & MESSAGE HANDLERS
// ============================================================

// processes user commands like /join, /part
func handleUserCommand(client *IRCClient, input string) {

	parts := strings.Fields(input)
	if len(parts) == 0 {
		return
	}

	command := strings.ToLower(parts[0])

	switch command {
	case "/join":
		if len(parts) < 2 {
			fmt.Println("Usage: /join #channel")
			return
		}

		channel := parts[1]

		//Ensure channel starts with #
		if !strings.HasPrefix(channel, "#") {
			channel = "#" + channel
		}

		//send JOIN command to server
		sendCommand(client.conn, fmt.Sprintf("JOIN %s", channel))
		fmt.Printf("Joining %s...\n", channel)

	case "/part":
		if client.currentChannel == "" {
			fmt.Println("You are not in a channel")
			return
		}

		//send PART command to server
		sendCommand(client.conn, fmt.Sprintf("PART %s", client.currentChannel))
		fmt.Printf("Leaving %s...\n", client.currentChannel)

	case "/nick":
		if len(parts) < 2 {
			fmt.Println("Usage: /nick <new_nickname>")
			return
		}
		newNick := parts[1]

		// Validate nickname
		if len(newNick) < 1 || len(newNick) > 16 {
			fmt.Println("❌ Nickname must be 1-16 characters")
			return
		}

		// Check for invalid characters
		for _, char := range newNick {
			if !isValidNickChar(char) {
				fmt.Println("❌ Nickname contains invalid characters")
				fmt.Println("Valid: letters, numbers, -, _, [, ], {, }, \\, |, `")
				return
			}
		}

		// Send NICK command to server
		sendCommand(client.conn, fmt.Sprintf("NICK %s", newNick))
		fmt.Printf("Changing nickname to %s...\n", newNick)

	case "/quit":
		// Get everything after /quit as the message
		input := strings.TrimSpace(input)
		quitMessage := strings.TrimPrefix(input, "/quit")
		quitMessage = strings.TrimSpace(quitMessage)
		
		// Use default if no message provided
		if quitMessage == "" {
			quitMessage = "Goodbye!"
		}

		fmt.Printf("Disconnecting with message: %s\n", quitMessage)

		quitCommand :=  fmt.Sprintf("QUIT :%s", quitMessage)
		sendCommand(client.conn, quitCommand)

		// Give server time to process before exiting
		time.Sleep(1 * time.Second)
		os.Exit(0)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Available: /join, /part, /nick, /quit")
	}
}

// handles different IRC message types
func handleMessage(client *IRCClient, msg Message, raw string) {

	switch msg.Command {
	case "PING":
		//Respond to PING with PONG
		if len(msg.Params) > 0 {
			pongMsg := fmt.Sprintf("PONG %s", msg.Params[0])
			sendCommand(client.conn, pongMsg)
		}

	case "001":
		//Welcome message - successfully connected
		fmt.Println("\n🎉 Successfully registered with IRC server!")
		if len(msg.Params) > 1 {
			fmt.Printf("Welcome: %s\n\n", msg.Params[1])
		}

		// CRITICAL FIX: Update nickname from server
		if len(msg.Params) > 0 {
			// First param is your actual nickname
			client.nickname = msg.Params[0]
			fmt.Printf("[INFO] Your nickname is: %s\n", client.nickname)
		}

	case "JOIN":
		//someone joined a channel
		if len(msg.Params) > 0 {
			channel := msg.Params[0]
			user := extractNick(msg.Prefix)

			if strings.EqualFold(user, client.nickname) {
				client.currentChannel = channel
				fmt.Printf("\n✅ You joined %s\n", channel)

				// Refresh prompt to show new channel
				fmt.Printf("[%s@%s] > ", client.nickname, client.currentChannel)
			} else {
				fmt.Printf("→ %s joined %s\n", user, channel)
			}
		}

	case "PART":
		//someone left a channel
		if len(msg.Params) > 0 {
			channel := msg.Params[0]
			user := extractNick(msg.Prefix)

			if user == client.nickname {
				// We left the channel
				client.currentChannel = ""
				fmt.Printf("\n✅ You left %s\n", channel)
			} else {
				// Someone else left
				fmt.Printf("← %s left %s\n", user, channel)
			}
		}

	case "NICK":
		// Someone changed their nickname
		if len(msg.Params) > 0 {
			oldNick := extractNick(msg.Prefix)
			newNick := msg.Params[0]

			if oldNick == client.nickname {
				// We changed our nickname
				client.nickname = newNick
				fmt.Printf("\n✅ You are now known as %s\n", newNick)
			} else {
				// Someone else changed nickname
				fmt.Printf("📝 %s is now known as %s\n", oldNick, newNick)
			}
		}

	case "PRIVMSG":
		// Someone sent a message to channel or us
		if len(msg.Params) >= 2 {
			sender := extractNick(msg.Prefix)
			target := msg.Params[0]
			message := msg.Params[1]

			// Display the message
			if strings.HasPrefix(target, "#") {
				// Channel message
				fmt.Printf("[%s] %s: %s\n", target, sender, message)
			} else {
				// Private message to us
				fmt.Printf("[PM from %s]: %s\n", sender, message)
			}

			// Re-display prompt for better UX
			if client.currentChannel != "" {
				fmt.Printf("[%s@%s] > ", client.nickname, client.currentChannel)
			} else {
				fmt.Printf("[%s] > ", client.nickname)
			}
		}

	case "QUIT":
		// Someone quit IRC
		if msg.Prefix != "" {
			user := extractNick(msg.Prefix)

			// Get quit message if provided
			var quitMsg string
			if len(msg.Params) > 0 {
				quitMsg = msg.Params[0]
			}

			// Display quit notification
			if quitMsg != "" {
				fmt.Printf("← %s has quit IRC (%s)\n", user, quitMsg)
			} else {
				fmt.Printf("← %s has quit IRC\n", user)
			}

			// Re-display prompt
			if client.currentChannel != "" {
				fmt.Printf("[%s@%s] > ", client.nickname, client.currentChannel)
			} else {
				fmt.Printf("[%s] > ", client.nickname)
			}
		}

	case "353":
		//Names list (users in channel)
		if len(msg.Params) >= 4 {
			channel := msg.Params[2]
			users := msg.Params[3]
			fmt.Printf("Users in %s: %s\n", channel, users)
		}

	case "432":
		// Erroneous nickname
		fmt.Println("❌ ERROR: Invalid nickname format")

	case "433":
		// Nickname already in use
		fmt.Println("❌ ERROR: Nickname is already in use!")
		fmt.Println("Please change NICKNAME in the code and restart.")

	case "436":
		// Nickname collision
		fmt.Println("❌ ERROR: Nickname collision")

	case "002", "003", "004":
		//server info messages
		if len(msg.Params) > 1 {
			fmt.Printf("Info: %s\n", msg.Params[1])
		}

	case "375":
		//start of MOTD(Message of The Day)
		fmt.Println("\n--- Message of the Day ---")

	case "372":
		//MOTD line
		if len(msg.Params) > 1 {
			fmt.Printf("%s\n", msg.Params[1])
		}

	case "376":
		//End of MOTD
		fmt.Println("--- End of Message of the Day")
		fmt.Println()

	case "NOTICE":
		//server notices
		if len(msg.Params) > 1 {
			fmt.Printf("Notice: %s\n", msg.Params[1])
		}

	default:
		// Handle unknown/unimplemented commands
		// This helps us see what messages we're not handling yet
		fmt.Printf("[Unhandled: %s] %s\n", msg.Command, raw)
	}
}

// ============================================================
// UTILITIES
// ============================================================

// sendCommand sends a command to the IRC server
func sendCommand(conn net.Conn, command string) {

	//IRC messages must end with \r\n
	_, err := conn.Write([]byte(command + "\r\n"))
	if err != nil {
		log.Println("Error sending command:", err)
	}
}

// parseMessage parses an IRC message into its components
func parseMessage(raw string) Message {

	msg := Message{}

	//check if message has a prefix(starts with :)
	if strings.HasPrefix(raw, ":") {
		//Extract prefix
		parts := strings.SplitN(raw[1:], " ", 2)
		if len(parts) < 2 {
			return msg
		}

		msg.Prefix = parts[0]
		raw = parts[1]
	}

	//split remaining message into command and params
	parts := strings.Split(raw, " ")
	if len(parts) == 0 {
		return msg
	}

	msg.Command = parts[0]

	//Extract parameters
	for i := 1; i < len(parts); i++ {
		//if parameter starts with ":", rest is trailing parameter
		if strings.HasPrefix(parts[i], ":") {
			//join remaining parts as one parameter
			trailing := strings.Join(parts[i:], " ")[1:] //remove leading ":"
			msg.Params = append(msg.Params, trailing)
			break
		}

		msg.Params = append(msg.Params, parts[i])
	}

	return msg
}

// extarct nickname from prefix
// Example: "john!~user@host" -> "john"
func extractNick(prefix string) string {

	if prefix == "" {
		return ""
	}

	//split at "!" to get nickname
	parts := strings.Split(prefix, "!")
	return parts[0]
}

// checks if a character is valid in IRC nicknames
func isValidNickChar(c rune) bool {
	// Valid nickname characters: A-Z a-z 0-9 - _ [ ] { } \ | `
	if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
		return true
	}
	validSpecial := "-_[]{}\\|`"
	for _, valid := range validSpecial {
		if c == valid {
			return true
		}
	}
	return false
}
