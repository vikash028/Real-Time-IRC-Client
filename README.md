# Real-Time IRC Client

A fully-functional Internet Relay Chat (IRC) client built in Go, implementing RFC 2812 IRC Client Protocol with concurrent goroutines and real-time messaging.

## 🚀 Features

- RFC 2812 IRC Client Protocol compliant
- Concurrent architecture with goroutines
- Real-time messaging via PRIVMSG
- Multi-channel support with JOIN/PART
- Nickname management with /nick command
- Thread-safe state management
- Automatic PING/PONG keepalive
- Custom quit messages
- Full IRC message parser

## 🛠️ Tech Stack

**Language:** Go 1.21+ | **Protocol:** IRC (RFC 2812) | **Transport:** TCP/IP | **Concurrency:** Goroutines & Channels

## 🚦 Getting Started

```bash
git clone https://github.com/vikash028/Real-Time-IRC-Client.git
cd Real-Time-IRC-Client
go run main.go
```

### Configuration

Edit constants in `main.go`:
```go
const (
    SERVER   = "chat.freenode.net:6667"
    NICKNAME = "CCClient123"
    USERNAME = "guest"
    REALNAME = "Coding Challenges Client"
)
```

## 📋 Commands

| Command | Description | Example |
|---------|-------------|---------|
| `/join #channel` | Join a channel | `/join #golang` |
| `/part` | Leave current channel | `/part` |
| `/nick <name>` | Change nickname | `/nick JohnDoe` |
| `/quit [msg]` | Disconnect | `/quit Goodbye!` |
| `<message>` | Send to channel | `Hello everyone!` |

## 🎯 Usage Example

```bash
[CCClient123] > /join #test
✅ You joined #test
Users in #test: @CCClient123 alice bob

[CCClient123@#test] > Hello everyone!
[#test] CCClient123: Hello everyone!
[#test] alice: Hi there!

[CCClient123@#test] > /nick JohnDoe
✅ You are now known as JohnDoe

[JohnDoe@#test] > /quit Thanks!
Disconnecting with message: Thanks!
```

## 🏗️ Architecture

**Two-Goroutine Design:**
- **Main Goroutine:** Reads user input, parses commands, sends to server
- **Reader Goroutine:** Receives server messages, parses IRC protocol, updates state

**Message Format (RFC 2812):**
```
[:prefix] command [params] [:trailing]
Example: :alice!~user@host PRIVMSG #test :Hello world
```

**Key Components:**
```go
// Registration
sendCommand(conn, "NICK CCClient123")
sendCommand(conn, "USER guest 0 * :Coding Challenges Client")

// Joining and messaging
sendCommand(conn, "JOIN #test")
sendCommand(conn, "PRIVMSG #test :Hello!")

// Keepalive
case "PING":
    sendCommand(conn, fmt.Sprintf("PONG %s", msg.Params[0]))
```

## 📊 Supported IRC Commands

**Client → Server:** NICK, USER, JOIN, PART, PRIVMSG, QUIT, PONG

**Server → Client (Handled):** 001-004 (welcome/info), 353 (user list), 432-433-436 (nickname errors), PING, JOIN, PART, NICK, PRIVMSG, QUIT, NOTICE

## 🧪 Testing

**Multiple Clients:**
```bash
# Terminal 1
go run main.go
[CCClient123] > /join #test

# Terminal 2  
go run main.go
[CCClient456] > /join #test
```

**Web Client:** Test with [Freenode Webchat](https://webchat.freenode.net/)

## 🎓 What I Learned

- Network protocol implementation (RFC 2812)
- Concurrent programming with goroutines and channels
- TCP/IP socket programming and connection management
- Text protocol parsing with structured message formats
- Thread-safe state management across goroutines
- Real-time systems with keepalive mechanisms

## 🔮 Future Enhancements

- SSL/TLS support (port 6697)
- SASL authentication
- Multiple channel tabs
- Message logging
- DCC file transfer
- Channel operator commands

## 📚 Resources

- [RFC 2812 - IRC Client Protocol](https://datatracker.ietf.org/doc/html/rfc2812)
- [RFC 2810-2813 - IRC Specifications](https://datatracker.ietf.org/doc/html/rfc2810)
- [Freenode Network](https://freenode.net/)

## 🙏 Acknowledgments

Built as part of [Coding Challenges #16](https://codingchallenges.substack.com/p/coding-challenge-16-an-irc-client) by John Crickett.

---

⭐ **If you found this helpful, give it a star!**

**Author:** Vikash - [GitHub](https://github.com/vikash028)
