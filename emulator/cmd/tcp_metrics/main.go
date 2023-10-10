package main

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/writer"
)

// sudo ss -it4 | grep "cubic|reno" | awk '{print $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18$19,$20,$21,$22,$23$24,$25$26,$27,$28,$29,$30,$31,$32,$33,$34}'

type SocketStatsTCP struct {
	CongestionAlgorithm   string  `json:"congestion_algorithm" parquet:"name=congestion_algorithm, type=BYTE_ARRAY"`
	WscaleSND             int32   `json:"wscalesnd" parquet:"name=window_scale_send, type=INT32, convertedtype=INT_32"`
	WscaleRCV             int32   `json:"wscalercv" parquet:"name=window_scale_receive, type=INT32, convertedtype=INT_32"`
	ReTransmissionTimeout int32   `json:"rto" parquet:"name=retransmission_timeout, type=INT32, convertedtype=INT_32"`  // unit m
	RTTMean               float32 `json:"rtt_mean" parquet:"name=rtt_mean, type=FLOAT"`                                 // unit m
	RTTVar                float32 `json:"rtt_var" parquet:"name=rtt_var, type=FLOAT"`                                   // unit m
	AckTimeout            int32   `json:"ato" parquet:"name=acknowledgement_timeout, type=INT32, convertedtype=INT_32"` // unit m
	PMTU                  int32   `json:"pmtu" parquet:"name=path_mtu, type=INT32, convertedtype=INT_32"`               // path MT
	RCVMSS                int32   `json:"rcvmss" parquet:"name=receive_maximum_segment_size, type=INT32, convertedtype=INT_32"`
	ADVMSS                int32   `json:"advmss" parquet:"name=advertised_maximum_segment_size, type=INT32, convertedtype=INT_32"` // Advertised MS
	MSS                   int32   `json:"mss" parquet:"name=maximum_segment_size, type=INT32, convertedtype=INT_32"`
	CWND                  int32   `json:"cwnd" parquet:"name=congestion_window, type=INT32, convertedtype=INT_32"`
	BytesSent             int64   `json:"bytes_sent" parquet:"name=bytes_sent, type=INT64, convertedtype=INT_64"`
	BytesAcked            int64   `json:"bytes_acked" parquet:"name=bytes_acked, type=INT64, convertedtype=INT_64"`
	BytesReceived         int64   `json:"bytes_received" parquet:"name=bytes_received, type=INT64, convertedtype=INT_64"`
	BytesRetrans          int64   `json:"bytes_retrans" parquet:"name=bytes_retrans, type=INT64, convertedtype=INT_64"`
	SegsOut               int64   `json:"segs_out" parquet:"name=segments_out, type=INT64, convertedtype=INT_64"`
	SegsIn                int64   `json:"segs_in" parquet:"name=segments_in, type=INT64, convertedtype=INT_64"`
	DataSegsOut           int64   `json:"data_segs_out" parquet:"name=data_segments_out, type=INT64, convertedtype=INT_64"`
	DataSegsIn            int64   `json:"data_segs_in" parquet:"name=data_segments_in, type=INT64, convertedtype=INT_64"`
	SendRate              float32 `json:"send" parquet:"name=send_rate, type=FLOAT"`
	LastSND               int64   `json:"lastsnd" parquet:"name=last_send, type=INT64, convertedtype=INT_64"`
	LastRCV               int64   `json:"lastrcv" parquet:"name=last_receive, type=INT64, convertedtype=INT_64"`
	LastAck               int64   `json:"lastack" parquet:"name=last_acknowledgment, type=INT64, convertedtype=INT_64"`
	PacingRate            float32 `json:"pacing_rate" parquet:"name=pacing_rate, type=FLOAT"`
	DeliveryRate          float32 `json:"delivery_rate" parquet:"name=delivery_rate, type=FLOAT"`
	Delivered             int64   `json:"delivered" parquet:"name=delivered, type=INT64, convertedtype=INT_64"`
	Busy                  int32   `json:"busy" parquet:"name=busy, type=INT32, convertedtype=INT_32"`
	RCVSpace              int32   `json:"rcv_space" parquet:"name=receive_space, type=INT32, convertedtype=INT_32"`
	RCVSSThresh           int32   `json:"rcv_ssthresh" parquet:"name=receive_slow_start_threshold, type=INT32, convertedtype=INT_32"`
	MINRTT                float32 `json:"minrtt" parquet:"name=minimum_rtt, type=FLOAT"`
	SNDWND                int32   `json:"snd_wnd" parquet:"name=send_window, type=INT32, convertedtype=INT_32"`
	RCV_RTT               float32 `json:"rcvrtt" parquet:"name=receive_rtt, type=FLOAT"`
	AppLimited            bool    `json:"app_limited" parquet:"name=application_limited, type=BOOLEAN"`
	Backoff               int32   `json:"backoff" parquet:"name=backoff, type=INT32, convertedtype=INT_32"`
	Unacked               int32   `json:"unacked" parquet:"name=unacked, type=INT32, convertedtype=INT_32"`
	DSACKDUPS             int32   `json:"dsack_dups" parquet:"name=dsack_dups, type=INT32, convertedtype=INT_32"`
	Lost                  int32   `json:"lost" parquet:"name=lost, type=INT32, convertedtype=INT_32"`
	Retrans               int32   `json:"retrans" parquet:"name=retrans, type=INT32, convertedtype=INT_32"`
	RetransTotal          int32   `json:"retrans_total" parquet:"name=retrans_total, type=INT32, convertedtype=INT_32"`
	Timestamp             int64   `parquet:"name=timestamp, type=INT64, convertedtype=TIMESTAMP_MILLIS"`
	Id                    int32   `parquet:"name=id, type=INT32, convertedtype=INT_32"`
}

const (
	UNDEFINED = iota
	TYPE_FLAG
	TYPE_FLOAT_PAIR
	TYPE_INT_PAIR_COMMA
	TYPE_INT_PAIR_SLASH
	TYPE_KV_INT
	TYPE_KV_FLOAT
	TYPE_KV_RATE
)

func preprocessRaw(input string) string {
	text := input
	text = strings.TrimRight(text, "\n ")
	text = strings.TrimSpace(text)
	text = strings.Replace(text, "send ", "send:", 1)
	text = strings.Replace(text, "pacing_rate ", "pacing_rate:", 1)
	text = strings.Replace(text, "delivery_rate ", "delivery_rate:", 1)
	// text = strings.Replace(text, "delivery_rate ", "delivery_rate:", 1)
	return text
}

func detectFormatting(text string) int {
	if !strings.Contains(text, ":") {
		return TYPE_FLAG
	} else if strings.Contains(text, "/") {
		if strings.Contains(text, ".") {
			return TYPE_FLOAT_PAIR
		} else {
			return TYPE_INT_PAIR_SLASH
		}
	} else if strings.Contains(text, ",") {
		return TYPE_INT_PAIR_COMMA
	} else {
		if strings.Contains(text, "bps") {
			return TYPE_KV_RATE
		} else if strings.Contains(text, ".") {
			return TYPE_KV_FLOAT
		} else {
			return TYPE_KV_INT
		}
	}

}

var congestion_algorithms = []string{"cubic", "reno", "bbr"}

// Suffix ms µs s
// Suffix kbps Mbps Gbps Tbps

// cubic wscale:8,7 rto:326 rtt:124.367/17.253 ato:40 mss:1368 pmtu:1500 rcvmss:1368 advmss:1448 cwnd:10 bytes_sent:32608 bytes_acked:32609 bytes_received:7109 segs_out:40 segs_in:34 data_segs_out:29 data_segs_in:13 send:880kbps lastsnd:49734 lastrcv:49568 lastack:49568 pacing_rate:1.76Mbps delivery_rate:290kbps delivered:30 app_limited busy:1424ms rcv_space:14480 rcv_ssthresh:64088 minrtt:113.006 snd_wnd:101888
// sudo ss -it4 | grep cubic | awk '{print $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18":"$19,$20,$21,$22,$23":"$24,$25":"$26,$27,$28,$29,$30,$31,$32,$33,$34}' | go run main.go
func parseColumn(col string, m map[string]any) error {
	format := detectFormatting(col)
	switch format {
	case TYPE_FLAG:
		k := trimKey(col)
		for _, congestion_algo := range congestion_algorithms {
			if k == congestion_algo {
				m["congestion_algorithm"] = k
				return nil
			}
		}
		m[k] = true
	case TYPE_FLOAT_PAIR:
		k, v1, v2, err := splitKVFloatPair(col)
		if err != nil {
			log.Println(k, v1, v2, err, col)
			return err
		}
		if k == "rtt" {
			m[k+"_mean"] = v1
			m[k+"_var"] = v2
		} else if k == "retrans" {
			m[k+"_01"] = v1
			m[k+"_02"] = v2
		} else {
			log.Fatal(k, v1, v2, "unsupported kv float pair")
		}
	case TYPE_INT_PAIR_COMMA:
		k, v1, v2, err := splitKVIntPair(col)
		if err != nil {
			log.Println(k, v1, v2, err, col)
			return err
		}
		if k == "wscale" {
			m[k+"snd"] = v1
			m[k+"rcv"] = v2
		} else {
			log.Fatal(k, v1, v2, "unsupported kv int pair")
		}
	case TYPE_INT_PAIR_SLASH:
		k, v1, v2, err := splitKVIntPairSlash(col)
		if err != nil {
			log.Println(k, v1, v2, err, col)
			return err
		}
		if k == "retrans" {
			m[k] = v1
			m[k+"total"] = v2
		} else {
			log.Fatal(k, v1, v2, "unsupproted kv int slash pair")
		}
	case TYPE_KV_INT:
		k, v, err := splitKVInt(col)
		if err != nil {
			log.Println(k, v, err, col)
			return err
		}
		m[k] = v
	case TYPE_KV_FLOAT:
		k, v, err := splitKVFloat(col)
		if err != nil {
			log.Println(k, v, err, col)
			return err
		}
		m[k] = v
	case TYPE_KV_RATE:
		k, v, err := parseRate(col)
		if err != nil {
			log.Println(k, v, err, col)
			return err
		}
		m[k] = v
	}
	return nil
}

func parseStats(columns []string) (res SocketStatsTCP, err error) {
	// log.Println(columns)
	// log.Println(len(columns))
	m := make(map[string]any)
	for _, col := range columns {
		// log.Println(i, col)
		err := parseColumn(col, m)
		if err != nil {
			log.Println(err, columns)
			return res, err
		}
	}
	jsonbody, err := json.Marshal(m)
	if err != nil {
		log.Fatal(err, columns)
	}

	if err := json.Unmarshal(jsonbody, &res); err != nil {
		log.Fatal(err, columns)
	}
	return res, nil

}

func main() {
	// f, err := os.Create("tcp_data")
	// tcpdataFile, err := ioutil.TempFile(os.TempDir(), "tcp_statistics"+time.Now().Format(time.Kitchen))
	// log.Println(tcpdataFile.Name())
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer tcpdataFile.Sync()
	// defer tcpdataFile.Close()
	var err error
	fw, err := local.NewLocalFileWriter("tcp_statistics" + time.Now().Format(time.Kitchen))
	if err != nil {
		log.Fatal("Can't create local file", err)
		return
	}

	//write
	pw, err := writer.NewParquetWriter(fw, new(SocketStatsTCP), 2)
	if err != nil {
		log.Println("Can't create parquet writer", err)
		return
	}
	pw.RowGroupSize = 1 * 256 * 1024 //256K
	pw.PageSize = 2 * 1024           //2K
	pw.CompressionType = parquet.CompressionCodec_SNAPPY

	Stop := func() {
		pw.Flush(true)
		if err = pw.WriteStop(); err != nil {
			log.Println("WriteStop error", err)
			return
		}
		fw.Close()
	}
	defer Stop()

	// scanner := bufio.NewScanner(os.Stdin)
	command_string := "ss -OHnit4 | grep -E \"(reno)|(cubic)\"" // | awk '{print $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18\":\"$19,$20,$21,$22,$23\":\"$24,$25\":\"$26,$27,$28,$29,$30,$31,$32,$33,$34}'
	argstr := []string{
		"/bin/bash",
		"-c",
		command_string,
	}
	interruptSignal := make(chan os.Signal, 1)
	signal.Notify(interruptSignal, syscall.SIGINT)
	var prevStats SocketStatsTCP = SocketStatsTCP{}
	// log.Println(cmd.Path, cmd.Args)
	// TODO proper sleep
	for counter := 0; counter < 20000; counter++ {
		select {
		case stopsignal := <-interruptSignal:
			log.Println(stopsignal)
			return
		default:
		}
		cmd := exec.Command(argstr[0], argstr[1:]...)
		// var out bytes.Buffer
		// cmd.Stdout = &out
		output, err := cmd.Output()
		if err != nil {
			log.Fatal(err)
		}

		text := string(output)
		// log.Print(text)
		lines := strings.Split(text, "\n")
		for id, line := range lines {
			if len(line) == 0 {
				continue
			}
			text := preprocessRaw(line)
			// text = strings.TrimRight(text, "\n ") // Trim newline and trailing whitespace
			columns := strings.Split(text, " ")
			if len(columns) < 5 {
				log.Println(columns)
				log.Println(line)
				continue
			}
			stats, err := parseStats(columns[5:])
			if err != nil {
				log.Fatalf("failed to parse stats:\n line:%s\ntext:%s\ncolumns:%v\n,stats%v", line, text, columns[5:], stats)
			}

			stats.Timestamp = time.Now().UnixMilli()
			stats.Id = int32(id)
			// log.Printf("id:%d cwnd %d, RTT[µ,σ]%f, %f", id, stats.CWND, stats.RTTMean, stats.RTTVar)
			if err = pw.Write(stats); err != nil {
				log.Println("Write error", err, stats, prevStats)
			} else {
				prevStats = stats
			}
			if counter%500 == 0 {
				pw.Flush(true)
			}
		}
		time.Sleep(50 * time.Millisecond)
	}

}

func trimKey(text string) string {
	return strings.TrimRight(text, " \t")
}

func removeUnits(text string) (prefix string, scale float64) {
	if strings.HasSuffix(text, "ms") {
		return strings.TrimSuffix(text, "ms"), 1
	} else if strings.HasSuffix(text, "kbps") {
		return strings.TrimSuffix(text, "kbps"), 1e3
	} else if strings.HasSuffix(text, "Mbps") {
		return strings.TrimSuffix(text, "Mbps"), 1e6
	} else if strings.HasSuffix(text, "Gbps") {
		return strings.TrimSuffix(text, "Gbps"), 1e9
	} else if strings.HasSuffix(text, "Tbps") {
		return strings.TrimSuffix(text, "Tbps"), 1e12
	} else if strings.HasSuffix(text, "bps") {
		return strings.TrimSuffix(text, "bps"), 1
	}
	if strings.ContainsAny(text, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		log.Println("error finding suffix in string", text)
	}
	return text, 1
}

func splitKVInt(kv string) (key string, value int, err error) {
	substrings := strings.Split(kv, ":")
	key = substrings[0]
	prefix, scale := removeUnits(substrings[1])

	value, err = strconv.Atoi(prefix)
	value = int(float64(value) * scale)

	if err != nil {
		// log.Fatal(err, kv)
		return key, value, err
	}
	return key, value, nil
}

func splitKVIntPair(kv string) (key string, v1 int, v2 int, err error) {
	substrings := strings.Split(kv, ":")
	key = substrings[0]
	pair := strings.Split(substrings[1], ",")
	v1, err = strconv.Atoi(pair[0])
	if err != nil {
		// log.Fatal(err, kv)
		return key, 0, 0, err
	}
	v2, err = strconv.Atoi(pair[1])
	if err != nil {
		return key, 0, 0, err
		// log.Fatal(err, kv)
	}
	return key, v1, v2, nil
}

// TODO only differnce is delimter between these 2 functions
func splitKVIntPairSlash(kv string) (key string, v1 int, v2 int, err error) {
	substrings := strings.Split(kv, ":")
	key = substrings[0]
	pair := strings.Split(substrings[1], "/")
	v1, err = strconv.Atoi(pair[0])
	if err != nil {
		// log.Fatal(err, kv)
		return key, 0, 0, err
	}
	v2, err = strconv.Atoi(pair[1])
	if err != nil {
		return key, 0, 0, err
		// log.Fatal(err, kv)
	}
	return key, v1, v2, nil
}

func splitKVFloat(kv string) (key string, value float32, err error) {
	substrings := strings.Split(kv, ":")
	key = substrings[0]
	prefix, scale := removeUnits(substrings[1])

	tvalue, err := strconv.ParseFloat(prefix, 32)
	if err != nil {
		return key, 0, nil
		// log.Fatal(err, kv)
	}
	return key, float32(float64(tvalue) * scale), nil
}

func splitKVFloatPair(kv string) (key string, v1, v2 float32, err error) {
	substrings := strings.Split(kv, ":")
	key = substrings[0]
	pair := strings.Split(substrings[1], "/")

	tv1, err := strconv.ParseFloat(pair[0], 32)
	if err != nil {
		// log.Fatal(err, kv)
		return key, 0, 0, err
	}
	tv2, err := strconv.ParseFloat(pair[1], 32)
	if err != nil {
		log.Fatal(err, kv)
		return key, 0, 0, err
	}
	return key, float32(tv1), float32(tv2), nil
}

func parseRate(rawrate string) (key string, value float32, err error) {
	// "delivery_rate210Gbps"
	substrings := strings.Split(rawrate, ":")
	key = substrings[0]

	prefix, scale := removeUnits(substrings[1])

	tv1, err := strconv.ParseFloat(prefix, 32)
	if err != nil {
		return key, 0, err
	}

	return key, float32(float64(tv1) * scale), nil
}

// res.CongestionAlgorithm = columns[0]
// _, v1, v2 := splitKVIntPair(columns[1])
// res.WscaleSND = v1
// res.WscaleRCV = v2
// log.Println(1)
// _, res.ReTransmissionTimeout = splitKVInt(columns[2])
// _, res.RTTMean, res.RTTVar = splitKVFloatPair(columns[3])
// _, res.AckTimeout = splitKVInt(columns[4])
// _, res.MaxSegmentSize = splitKVInt(columns[5])
// _, res.PMTU = splitKVInt(columns[6])
// log.Println(2)
// _, res.RCVMSS = splitKVInt(columns[7])
// _, res.ADVMSS = splitKVInt(columns[8])
// _, res.CWND = splitKVInt(columns[9])
// _, res.BytesSent = splitKVInt(columns[10])
// _, res.BytesAcked = splitKVInt(columns[11])
// _, res.BytesReceived = splitKVInt(columns[12])
// log.Println(3)
// _, res.SegsOut = splitKVInt(columns[13])
// _, res.SegsIn = splitKVInt(columns[14])
// _, res.DataSegsOut = splitKVInt(columns[15])
// _, res.DataSegsIn = splitKVInt(columns[16])
// _, res.SendRate = parseRate(columns[17])
// _, res.LastSND = splitKVInt(columns[18])
// _, res.LastRCV = splitKVInt(columns[19])
// log.Println(4)
// _, res.LastAck = splitKVInt(columns[20])
// _, res.PacingRate = parseRate(columns[21])
// _, res.DeliveryRate = parseRate(columns[22])
// _, res.Delivered = splitKVInt(columns[23])
// log.Println(5)
// if len(columns) == 30 {
// 	log.Println(6)
// 	_, res.Busy = splitKVInt(columns[25])
// 	_, res.RCVSpace = splitKVInt(columns[26])
// 	_, res.RCVSSTHresh = splitKVInt(columns[27])
// 	_, res.MINRTT = splitKVFloat(columns[28])
// 	_, res.SNDWND = splitKVInt(columns[29])
// } else {
// 	log.Println(7)
// 	_, res.Busy = splitKVInt(columns[25-1])
// 	_, res.RCVSpace = splitKVInt(columns[26-1])
// 	_, res.RCVSSTHresh = splitKVInt(columns[27-1])
// 	_, res.MINRTT = splitKVFloat(columns[28-1])
// 	_, res.SNDWND = splitKVInt(columns[29-1])
// }
