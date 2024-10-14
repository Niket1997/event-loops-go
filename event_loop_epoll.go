package main

//var (
//	host       = "127.0.0.1" // Server IP address
//	port       = 8080        // Server port
//	maxClients = 20000       // Maximum number of concurrent clients
//)

//func RunAsyncTCPServerEpoll() error {
//	log.Printf("starting an asynchronous TCP server on %s:%d", host, port)
//
//	// Create a socket
//	serverFD, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
//	if err != nil {
//		return fmt.Errorf("socket creation failed: %v", err)
//	}
//	defer unix.Close(serverFD)
//
//	// Set the socket to non-blocking mode
//	if err := unix.SetNonblock(serverFD, true); err != nil {
//		return fmt.Errorf("failed to set non-blocking mode: %v", err)
//	}
//
//	// Allow address reuse
//	if err := unix.SetsockoptInt(serverFD, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1); err != nil {
//		return fmt.Errorf("failed to set SO_REUSEADDR: %v", err)
//	}
//
//	// Bind the IP & the port
//	addr := &unix.SockaddrInet4{Port: port}
//	copy(addr.Addr[:], net.ParseIP(host).To4())
//	if err := unix.Bind(serverFD, addr); err != nil {
//		return fmt.Errorf("failed to bind socket: %v", err)
//	}
//
//	// Start listening
//	if err := unix.Listen(serverFD, maxClients); err != nil {
//		return fmt.Errorf("failed to listen on socket: %v", err)
//	}
//
//	// Create epoll instance
//	epfd, err := unix.EpollCreate1(0)
//	if err != nil {
//		return fmt.Errorf("failed to create epoll instance: %v", err)
//	}
//	defer unix.Close(epfd)
//
//	// Register the serverFD with epoll
//	event := &unix.EpollEvent{
//		Events: unix.EPOLLIN,
//		Fd:     int32(serverFD),
//	}
//	if err := unix.EpollCtl(epfd, unix.EPOLL_CTL_ADD, serverFD, event); err != nil {
//		return fmt.Errorf("failed to add server FD to epoll: %v", err)
//	}
//
//	// Event loop
//	events := make([]unix.EpollEvent, maxClients)
//	for {
//		nevents, err := unix.EpollWait(epfd, events, -1)
//		if err != nil {
//			if err == unix.EINTR {
//				continue // Interrupted system call, retry
//			}
//			return fmt.Errorf("epoll_wait error: %v", err)
//		}
//
//		for i := 0; i < nevents; i++ {
//			ev := events[i]
//			fd := int(ev.Fd)
//
//			if fd == serverFD {
//				// Accept the incoming connection from client
//				nfd, sa, err := unix.Accept(serverFD)
//				if err != nil {
//					if err == unix.EAGAIN || err == unix.EWOULDBLOCK {
//						continue // No more incoming connections
//					}
//					log.Printf("failed to accept connection: %v", err)
//					continue
//				}
//
//				// Set the new socket to non-blocking mode
//				if err := unix.SetNonblock(nfd, true); err != nil {
//					log.Printf("failed to set non-blocking mode on client FD: %v", err)
//					unix.Close(nfd)
//					continue
//				}
//
//				// Register the new client FD with epoll
//				clientEvent := &unix.EpollEvent{
//					Events: unix.EPOLLIN | unix.EPOLLET, // Use edge-triggered mode
//					Fd:     int32(nfd),
//				}
//
//				if err := unix.EpollCtl(epfd, unix.EPOLL_CTL_ADD, nfd, clientEvent); err != nil {
//					log.Printf("failed to add client FD to epoll: %v", err)
//					unix.Close(nfd)
//					continue
//				}
//
//				log.Printf("accepted new connection from %v", sa)
//			} else {
//				// Handle client I/O
//				for {
//					buf := make([]byte, 1024)
//					n, err := unix.Read(fd, buf)
//					if n > 0 {
//						// Process the data received
//						data := buf[:n]
//						log.Printf("received data from client FD %d: %s", fd, string(data))
//
//						// Echo the data back to the client (optional)
//						if _, err := unix.Write(fd, data); err != nil {
//							if err == unix.EAGAIN || err == unix.EWOULDBLOCK {
//								// Socket buffer is full, can't write now
//								continue
//							}
//							log.Printf("failed to write to client FD %d: %v", fd, err)
//							// Close the connection
//							unix.Close(fd)
//							break
//						}
//					}
//					if err != nil {
//						if err == unix.EAGAIN || err == unix.EWOULDBLOCK {
//							// No more data to read now
//							break
//						}
//						log.Printf("failed to read from client FD %d: %v", fd, err)
//						// Close the connection
//						unix.Close(fd)
//						break
//					}
//					if n == 0 {
//						// Connection closed by client
//						log.Printf("client FD %d closed the connection", fd)
//						unix.Close(fd)
//						break
//					}
//				}
//			}
//		}
//	}
//}
//
//func main() {
//	if err := RunAsyncTCPServerEpoll(); err != nil {
//		log.Fatalf("server error: %v", err)
//	}
//}
