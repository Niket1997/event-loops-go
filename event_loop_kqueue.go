package main

import (
	"fmt"
	"golang.org/x/sys/unix"
	"log"
	"net"
)

var (
	host       = "127.0.0.1" // Server IP address
	port       = 8080        // Server port
	maxClients = 20000       // Maximum number of concurrent clients
)

func RunAsyncTCPServerUnix() error {
	log.Printf("starting an asynchronous TCP server on %s:%d", host, port)

	// Create kqueue event objects to hold events
	events := make([]unix.Kevent_t, maxClients)

	// Create a socket
	serverFD, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, unix.IPPROTO_TCP)
	if err != nil {
		return fmt.Errorf("socket creation failed: %v", err)
	}
	defer unix.Close(serverFD)

	// Set the socket to non-blocking mode
	if err := unix.SetNonblock(serverFD, true); err != nil {
		return fmt.Errorf("failed to set non-blocking mode: %v", err)
	}

	// Allow address reuse
	if err := unix.SetsockoptInt(serverFD, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1); err != nil {
		return fmt.Errorf("failed to set SO_REUSEADDR: %v", err)
	}

	// Bind the IP & the port
	addr := &unix.SockaddrInet4{Port: port}
	copy(addr.Addr[:], net.ParseIP(host).To4())
	if err := unix.Bind(serverFD, addr); err != nil {
		return fmt.Errorf("failed to bind socket: %v", err)
	}

	// Start listening
	if err := unix.Listen(serverFD, maxClients); err != nil {
		return fmt.Errorf("failed to listen on socket: %v", err)
	}

	// Create kqueue instance
	kq, err := unix.Kqueue()
	if err != nil {
		return fmt.Errorf("failed to create kqueue: %v", err)
	}
	defer unix.Close(kq)

	// Register the serverFD with kqueue
	kev := unix.Kevent_t{
		Ident:  uint64(serverFD),
		Filter: unix.EVFILT_READ,
		Flags:  unix.EV_ADD,
	}

	if _, err := unix.Kevent(kq, []unix.Kevent_t{kev}, nil, nil); err != nil {
		return fmt.Errorf("failed to register server FD with kqueue: %v", err)
	}

	// Event loop
	for {
		nevents, err := unix.Kevent(kq, nil, events, nil)
		if err != nil {
			if err == unix.EINTR {
				continue // Interrupted system call, retry
			}
			return fmt.Errorf("kevent error: %v", err)
		}

		for i := 0; i < nevents; i++ {
			ev := events[i]
			fd := int(ev.Ident)

			if fd == serverFD {
				// Accept the incoming connection from client
				nfd, sa, err := unix.Accept(serverFD)
				if err != nil {
					log.Printf("failed to accept connection: %v", err)
					continue
				}

				// Set the new socket to non-blocking mode
				if err := unix.SetNonblock(nfd, true); err != nil {
					log.Printf("failed to set non-blocking mode on client FD: %v", err)
					unix.Close(nfd)
					continue
				}

				// Register the new client FD with kqueue
				clientKev := unix.Kevent_t{
					Ident:  uint64(nfd),
					Filter: unix.EVFILT_READ,
					Flags:  unix.EV_ADD,
				}

				if _, err := unix.Kevent(kq, []unix.Kevent_t{clientKev}, nil, nil); err != nil {
					log.Printf("failed to register client FD with kqueue: %v", err)
					unix.Close(nfd)
					continue
				}

				log.Printf("accepted new connection from %v", sa)
			} else {
				// Handle client I/O
				buf := make([]byte, 1024)
				n, err := unix.Read(fd, buf)
				if err != nil {
					if err == unix.EAGAIN || err == unix.EWOULDBLOCK {
						continue // No data available right now
					}
					log.Printf("failed to read from client FD %d: %v", fd, err)
					// Remove the FD from kqueue and close it
					kev := unix.Kevent_t{
						Ident:  uint64(fd),
						Filter: unix.EVFILT_READ,
						Flags:  unix.EV_DELETE,
					}
					unix.Kevent(kq, []unix.Kevent_t{kev}, nil, nil)
					unix.Close(fd)
					continue
				}

				if n == 0 {
					// Connection closed by client
					log.Printf("client FD %d closed the connection", fd)
					kev := unix.Kevent_t{
						Ident:  uint64(fd),
						Filter: unix.EVFILT_READ,
						Flags:  unix.EV_DELETE,
					}
					unix.Kevent(kq, []unix.Kevent_t{kev}, nil, nil)
					unix.Close(fd)
					continue
				}

				// Process the data received
				data := buf[:n]
				log.Printf("received data from client FD %d: %s", fd, string(data))

				// Echo the data back to the client (optional)
				if _, err := unix.Write(fd, data); err != nil {
					log.Printf("failed to write to client FD %d: %v", fd, err)
					// Handle write error if necessary
				}
			}
		}
	}
}

func main() {
	if err := RunAsyncTCPServerUnix(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
