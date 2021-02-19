package minisentinel

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/alicebob/miniredis/v2/server"
)

const msgInvalidSentinelCommand = "ERR unknown command '%s'"

func commandsSentinel(s *Sentinel) {
	s.srv.Register("SENTINEL", s.cmdsSentinel)
}

func initSentinelCmdHandler(s *Sentinel) {
	s.cmdHandler = map[string]func(*server.Peer, string, []string) error{
		"MASTERS":                 s.mastersCommand,
		"GET-MASTER-ADDR-BY-NAME": s.getMasterAddrByNameCommand,
		"SLAVES":                  s.slavesCommand,
		"SENTINELS":               s.sentinelsCommand,
	}
}

// cmdsSentinel - entry point for all commands that start with SENTINEL
func (s *Sentinel) cmdsSentinel(c *server.Peer, cmd string, args []string) {
	if !isSentinelCmd(cmd) {
		c.WriteError(fmt.Sprintf(msgInvalidSentinelCommand, cmd))
		return
	}
	if len(args) > 2 {
		c.WriteError(errWrongNumber(cmd))
		return
	}
	if !s.handleAuth(c) {
		return
	}
	subCmd := strings.ToUpper(args[0])

	cmdFn, ok := s.cmdHandler[subCmd]
	if !ok {
		c.WriteError(fmt.Sprintf(msgInvalidSentinelCommand, subCmd))
		return
	}
	err := cmdFn(c, cmd, args)
	if err != nil {
		c.WriteError(err.Error())
	}
}

func (s *Sentinel) getMasterAddrByNameCommand(c *server.Peer, cmd string, args []string) error {
	if !isSentinelCmd(cmd) {
		return fmt.Errorf(msgInvalidSentinelCommand, cmd)
	}
	subCmd := strings.ToUpper(args[0])
	if subCmd != "GET-MASTER-ADDR-BY-NAME" {
		return fmt.Errorf(msgInvalidSentinelCommand, subCmd)
	}
	if strings.ToUpper(s.masterInfo.Name) != strings.ToUpper(args[1]) {
		c.WriteLen(-1)
		return nil
	}
	c.WriteLen(2)
	c.WriteBulk(s.master.Host())
	c.WriteBulk(s.master.Port())
	return nil
}
func (s *Sentinel) slavesCommand(c *server.Peer, cmd string, args []string) error {
	if !isSentinelCmd(cmd) {
		return fmt.Errorf(msgInvalidSentinelCommand, cmd)
	}
	subCmd := strings.ToUpper(args[0])
	if subCmd != "SLAVES" {
		return fmt.Errorf(msgInvalidSentinelCommand, subCmd)
	}
	c.WriteLen(1)
	c.WriteLen(40)
	t := reflect.TypeOf(s.replicaInfo)
	v := reflect.ValueOf(s.replicaInfo)

	// Iterate over all available fields and read the tag value
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("mapstructure")
		c.WriteBulk(tag)
		c.WriteBulk(v.Field(i).Interface().(string))
	}

	return nil
}

func (s *Sentinel) mastersCommand(c *server.Peer, cmd string, args []string) error {
	if !isSentinelCmd(cmd) {
		return fmt.Errorf(msgInvalidSentinelCommand, cmd)
	}
	subCmd := strings.ToUpper(args[0])
	if subCmd != "MASTERS" {
		return fmt.Errorf(msgInvalidSentinelCommand, subCmd)
	}
	c.WriteLen(1)
	c.WriteLen(40)
	t := reflect.TypeOf(s.masterInfo)
	v := reflect.ValueOf(s.masterInfo)

	// Iterate over all available fields and read the tag value
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("mapstructure")
		c.WriteBulk(tag)
		c.WriteBulk(v.Field(i).Interface().(string))
	}

	return nil
}

func (s *Sentinel) sentinelsCommand(c *server.Peer, cmd string, args []string) error {
	if !isSentinelCmd(cmd) {
		return fmt.Errorf(msgInvalidSentinelCommand, cmd)
	}
	subCmd := strings.ToUpper(args[0])
	if subCmd != "SENTINELS" {
		return fmt.Errorf(msgInvalidSentinelCommand, subCmd)
	}
	sentinelInfo := s.SentinelInfo()
	t := reflect.TypeOf(sentinelInfo)
	v := reflect.ValueOf(sentinelInfo)

	c.WriteLen(1)
	c.WriteLen(t.NumField())

	// Iterate over all available fields and read the tag value
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("mapstructure")
		c.WriteBulk(tag)
		c.WriteBulk(v.Field(i).Interface().(string))
	}

	return nil
}

func isSentinelCmd(cmd string) bool {
	if strings.ToUpper(cmd) != "SENTINEL" {
		return false
	}
	return true
}
