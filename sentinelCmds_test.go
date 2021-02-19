package minisentinel

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/matryer/is"

	"github.com/gomodule/redigo/redis"
)

func testSetup(t *testing.T) (*miniredis.Miniredis, *Sentinel, redis.Conn) {
	is := is.New(t)
	m, err := miniredis.Run()
	is.NoErr(err)

	s := NewSentinel(m, WithReplica(m))
	is.Equal(s.Master(), m)  // make sure the master is correctly set
	is.Equal(s.Replica(), m) // make sure the replicas are correctly set

	err = s.Start()
	is.NoErr(err)

	c, err := redis.Dial("tcp", s.Addr())
	is.NoErr(err)
	return m, s, c
}
func TestSlaves(t *testing.T) {
	is := is.New(t)
	m, s, c := testSetup(t)
	defer m.Close()
	defer s.Close()
	// SLAVES command
	{
		// results is an []interfaces which points to [][]strings
		results, err := c.Do("SENTINEL", "SLAVES")
		is.NoErr(err)
		info, err := redis.Strings(results.([]interface{})[0], nil)
		sInfo, err := NewReplicaInfoFromStrings(info)
		is.NoErr(err)
		is.Equal(sInfo.Name, "mymaster")
		is.Equal(sInfo.Port, m.Port())
		is.Equal(sInfo.IP, m.Host())
	}
}
func TestMasters(t *testing.T) {
	is := is.New(t)
	m, s, c := testSetup(t)
	defer m.Close()
	defer s.Close()

	// MASTERS command
	{
		// results is an []interfaces which points to [][]strings
		results, err := c.Do("SENTINEL", "MASTERS")
		is.NoErr(err)
		info, err := redis.Strings(results.([]interface{})[0], nil)
		mInfo, err := NewMasterInfoFromStrings(info)
		is.NoErr(err)
		is.Equal(mInfo.Name, "mymaster")
		is.Equal(mInfo.Port, m.Port())
		is.Equal(mInfo.IP, m.Host())
	}
}
func TestGetMasterAddrByName(t *testing.T) {
	is := is.New(t)
	m, s, c := testSetup(t)
	defer m.Close()
	defer s.Close()
	// GET-MASTER-ADDR-BY-NAME command
	{
		results, err := redis.Strings(c.Do("SENTINEL", "GET-MASTER-ADDR-BY-NAME", "MYMASTER"))
		is.NoErr(err)
		t.Log(results)
		is.Equal(results[0], m.Host())
		is.Equal(results[1], m.Port())
	}
}
func TestSentinels(t *testing.T) {
	is := is.New(t)
	m, s, c := testSetup(t)
	defer m.Close()
	defer s.Close()
	// SENTINELS command
	{
		results, err := c.Do("SENTINEL", "SENTINELS", "MYMASTER")
		is.NoErr(err)
		info, err := redis.Strings(results.([]interface{})[0], nil)
		sInfo, err := NewSentinelInfoFromStrings(info)
		is.NoErr(err)
		is.Equal(sInfo.Name, "sentinel-mymaster")
		is.Equal(sInfo.Port, s.Port())
		is.Equal(sInfo.IP, m.Host())
		is.Equal(sInfo.Flags, "sentinel")
	}
}