package minisentinel

import "github.com/alicebob/miniredis/v2/server"

func commandsPing(s *Sentinel) {
	s.srv.Register("PING", s.cmdPing)
	s.srv.Register("AUTH", s.cmdAuth)
}

// cmdPing
func (s *Sentinel) cmdPing(c *server.Peer, cmd string, args []string) {
	if len(args) != 0 {
		c.WriteError(errWrongNumber(cmd))
		return
	}
	if !s.handleAuth(c) {
		return
	}
	c.WriteInline("PONG")
}

func (s *Sentinel) cmdAuth(c *server.Peer, cmd string, args []string) {
	if len(args) != 1 {
		c.WriteError(errWrongNumber(cmd))
		return
	}

	pw := args[0]

	s.Lock()
	defer s.Unlock()
	if s.password == "" {
		c.WriteError("ERR Client sent AUTH, but no password is set")
		return
	}
	if s.password != pw {
		c.WriteError("ERR invalid password")
		return
	}

	setAuthenticated(c)
	c.WriteOK()
}
