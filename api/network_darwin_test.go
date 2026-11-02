//go:build darwin

package api

import "testing"

func TestParseNetstat(t *testing.T) {
	const out = `Active Internet connections (including servers)
Proto Recv-Q Send-Q  Local Address          Foreign Address        (state)          rxbytes      txbytes  rhiwat  shiwat          process:pid    state  options           gencnt    flags   flags1 usecnt rtncnt fltrs
tcp4       0      0  10.1.20.90.51380       160.79.104.10.443      ESTABLISHED         4638         4482  131072  131768           claude:43321  00102 00000008 0000000000282f3c 20000081 04000900      3      0 000004
tcp6       0      0  *.51376                *.*                    LISTEN                 0            0  131072  131072         rapportd:753    00100 00000006 0000000000282b7e 00000001 00000800      1      0 000000
udp4       0      0  10.1.20.90.51363       192.178.183.94.443                        10400         5928 1048576   29040 Google Chrome He:79506  00102 00000000 000000000028303b 20000000 04200900      2      0 000002
`
	got := parseNetstat(out)
	want := []netSocket{
		{PID: 43321, Key: "tcp4|10.1.20.90.51380|160.79.104.10.443|43321", Rx: 4638, Tx: 4482},
		{PID: 753, Key: "tcp6|*.51376|*.*|753", Rx: 0, Tx: 0},
		{PID: 79506, Key: "udp4|10.1.20.90.51363|192.178.183.94.443|79506", Rx: 10400, Tx: 5928},
	}
	if len(got) != len(want) {
		t.Fatalf("parsed %d sockets, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("socket %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}
