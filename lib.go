//Local network IP crawler library, looks for a top-level README, and open port 80 and 443.
// TODO: Add documentation
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

// Helper function that increments IP addresses.
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// Main structure used for initiating a crawler.
type Crawler struct {
	Name           string
	LastIP         string
	FilesToLookFor []string
	IPChunkSize    int
	GoRoutines     int
}

// Uses the above struct and initiates it.
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

// Sends a request to said IP address + extension, if the response is 200 OK it will attempt to download the website.
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

// Function called via go routine. Returns results over a channel.
// Initiated as a go routine and returns various information back to the main program.
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

// Main function that does everything. I'll make these doc's way better when I'm not in Math class.
func (c Crawler) Discover() {
	var cidrToCrawl string
	lip := c.LastIP
	lookfor := c.FilesToLookFor
	chunksize := c.IPChunkSize
	routines := c.GoRoutines
	_ = lip
	// I though this would be implied but apparently not.
	if routines > chunksize {
		log.Fatal("Amount of `GoRoutines` should not be greater than `IPChunkSize`")
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	// Did you know this is a lot harder to recreate in C++?
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
	// On your horses...
	// Get ready!!
	// GO!!!
	// https://youtu.be/uUpDG680uew?t=4s
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
