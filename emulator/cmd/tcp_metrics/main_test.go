package main

import (
	"log"
	"strings"
	"testing"
)

var fixture1 string = "ESTAB    0      0      192.168.0.149:39714 162.159.135.234:443   reno wscale:13,7 rto:210 rtt:9.602/0.123 ato:62 mss:1368 pmtu:1500 rcvmss:1129 advmss:1448 cwnd:10 bytes_sent:24349 bytes_acked:24350 bytes_received:177761 segs_out:2511 segs_in:2512 data_segs_out:328 data_segs_in:2184 send 11397625bps lastsnd:1442 lastrcv:1337 lastack:1337 pacing_rate 22794064bps delivery_rate 3304680bps delivered:329 app_limited busy:3236ms rcv_rtt:286579 rcv_space:64494 rcv_ssthresh:523738 minrtt:9.287 snd_wnd:65536                                                   "

func TestParseLine(t *testing.T) {
	cases := []string{
		fixture1,
		// fixture2,
		// fixture3,
	}
	for i, testCase := range cases {
		t.Log(i)
		text := preprocessRaw(testCase)
		columns := strings.Split(text, " ")
		stats, err := parseStats(columns)
		if err != nil {
			t.Error(err, columns, text)
		} else {
			log.Print(stats.CWND)
		}

	}

}
