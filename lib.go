package discover

import (
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup

func inc(ip net.IP) {
	for l := len(ip) - 1; l >= 0; l-- {
		ip[l]++
		if ip[l] > 0 {
			break
		}
	}
}

type Crawler struct {
	Name           string
	LastIP         string
	FilesToLookFor []string
	IPChunkSize    int
	GoRoutines     int
}

func NewCrawler() *Crawler {
	c := new(Crawler)
	return c
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func Download(ip string, ext string) bool {
	timeout := time.Duration(1 * time.Second)
	netclient := http.Client{Timeout: timeout}
	response, err := netclient.Get("http://" + ip + "/" + ext)
	if err != nil {
		return false
	}
	if response.StatusCode != 200 {
		log.Printf("IP: %s\nProtocol: %s\nStatus: %s\n", ip, response.Proto, response.Status)
		return false
	}
	log.Printf("IP: %s\nProtocol: %s\nStatus: %s\n", ip, response.Proto, response.Status)
	exists, err := fileExists("data/")
	if err != nil {
		log.Fatal(err)
	}
	if exists != true {
		os.Mkdir("data/", 0700)
	}
	out, err := os.Create("data/" + ip)
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(out, response.Body)
	if err != nil {
		log.Fatal(err)
	}
	response.Body.Close()
	out.Close()
	return true
}

func discoverer(id int, start <-chan bool, jobs <-chan string, resblock <-chan bool, results chan<- string, lookfor []string) {
	var results_list []string
	<-start
	for j := range jobs {
		for _, ext := range lookfor {
			ip := j + ":443"
			response := Download(ip, ext)
			if response == true {
				results_list = append(results_list, ip)
			}
			//----------------------------------//
			ip = j + ":80"
			response = Download(ip, ext)
			if response == true {
				results_list = append(results_list, ip)
			}
		}
	}
	wg.Done()
}

func (c Crawler) Discover() {
	var cidrToCrawl string
	lip := c.LastIP
	lookfor := c.FilesToLookFor
	chunksize := c.IPChunkSize
	routines := c.GoRoutines
	_ = lip
	if routines > chunksize {
		log.Fatal("Amount of `GoRoutines` should not be greater than `IPChunkSize`")
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	for _, iface := range ifaces {
		addrs, ifaceErr := iface.Addrs()
		if ifaceErr != nil {
			log.Fatal(err)
		}
		for _, addr := range addrs {
			oip, _, cidrErr := net.ParseCIDR(addr.String())
			if cidrErr != nil {
				log.Fatal(err)
			}
			if str := strings.HasPrefix(addr.String(), "127"); str == false {
				if oip.To4() != nil {
					cidrToCrawl = addr.String()
				}
			}
		}
	}
	resultschan := make(chan string, chunksize*len(c.FilesToLookFor))
	jobschan := make(chan string)
	startblock := make(chan bool)
	resblock := make(chan bool)
	addr := cidrToCrawl
	for w := 1; w <= routines; w++ {
		wg.Add(1)
		go discoverer(w, startblock, jobschan, resblock, resultschan, lookfor)
	}
	oip, ipnet, err := net.ParseCIDR(addr)
	if err != nil {
		log.Fatal(err)
	}
	i := 0
	close(startblock)
	for ip := oip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		if i <= chunksize {
			jobschan <- ip.String()
			i++
			if i == chunksize {
				close(jobschan)
				break
			}
		}
	}
	wg.Wait()
	close(resultschan)
}
