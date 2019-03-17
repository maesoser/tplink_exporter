package tplink

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	LOGIN_URL       = "/userRpm/LoginRpm.htm?Save=Save"
	LOGOUT_URL      = "/userRpm/LogoutRpm.htm"
	WAN_TRAFFIC_URL = "/userRpm/StatusRpm.htm"
	CLIENTS_URL     = "/userRpm/AssignedIpAddrListRpm.htm"
	STATS_URL       = "/userRpm/SystemStatisticRpm.htm?itnerval=10&Num_per_page=100"
	REBOOT_URL      = "/userRpm/SysRebootRpm.htm?Reboot=Reboot"
	AUTH_KEY_RE     = "[0-9A-Za-z.]+/([A-Z]{16})/userRpm/Index.htm"
)

type Client struct {
	Name    string
	MAC     string
	Addr    string
	Lease   float64
	Packets float64
	Bytes   float64
}

type Router struct {
	Client  http.Client
	Cookie  http.Cookie
	Token   string
	Address string
	User    string
	Pass    string
}

func getMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func parseLease(LeaseTime string) float64 {
	LeaseTime = strings.Replace(LeaseTime, " ", "", -1)
	if LeaseTime == "Permanent" {
		return 0
	}
	timeArray := strings.Split(LeaseTime, ":")
	if len(timeArray) != 3 {
		return 0
	}
	h, err := strconv.Atoi(timeArray[0])
	if err != nil {
		fmt.Println(err)
	}
	m, _ := strconv.Atoi(timeArray[1])
	s, _ := strconv.Atoi(timeArray[2])
	total := h*3600 + m*60 + s
	return float64(total)
}

//Init configures the http client and generates the cookie
func (r *Router) Init() {
	hashpass := getMD5Hash(r.Pass)
	auth := base64.StdEncoding.EncodeToString([]byte(r.User + ":" + hashpass))
	r.Cookie = http.Cookie{Name: "Authorization", Value: auth}
	r.Client = http.Client{Timeout: time.Second * 2}
}

// Login retrieves the Token needed to perform requests to the router
func (r *Router) Login() error {
	req, err := http.NewRequest("GET", "http://"+r.Address+LOGIN_URL, nil)
	if err != nil {
		return err
	}
	req.AddCookie(&r.Cookie)
	response, err := r.Client.Do(req)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	expr, err := regexp.Compile(AUTH_KEY_RE)
	if err != nil {
		return err
	}
	matches := expr.FindAllStringSubmatch(string(body), -1)
	if len(matches) != 1 {
		return fmt.Errorf("Token not found on body:\n%s", string(body))
	}
	if len(matches[0]) != 2 {
		return fmt.Errorf("Token not found on body:\n%s", string(body))
	}
	r.Token = matches[0][1]
	return nil
}

//Get makes a get request adding the authentication needed.
func (r *Router) Get(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.AddCookie(&r.Cookie)
	req.Header.Set("Referer", url)
	response, err := r.Client.Do(req)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// GetWANTraffic obtains the router's total traffic metrics
func (r *Router) GetWANTraffic() (float64, float64, error) {
	body, err := r.Get("http://" + r.Address + "/" + r.Token + WAN_TRAFFIC_URL)
	if err != nil {
		return 0, 0, err
	}
	expr, err := regexp.Compile(`(?m)var statistList = new Array\(\n\"([^\"]*)\", \"([^\"]*)`)
	if err != nil {
		return 0, 0, err
	}
	matches := expr.FindAllStringSubmatch(body, -1)
	if len(matches) != 1 {
		return 0, 0, err
	}
	stats := matches[0]
	rx, _ := strconv.ParseFloat(strings.Replace(stats[1], ",", "", -1), 64)
	tx, _ := strconv.ParseFloat(strings.Replace(stats[2], ",", "", -1), 64)

	return tx / 1024, rx / 1024, nil
}

// GetClients obtain the list of clients connected to the router's wifi
func (r *Router) GetClients() ([]Client, error) {
	var clients []Client
	body, err := r.Get("http://" + r.Address + "/" + r.Token + CLIENTS_URL)
	if err != nil {
		return clients, err
	}
	expr, err := regexp.Compile(`(?m)(\"([^\"]*)\", \"([^\"]*)\", \"([^\"]*)\", \"([^\"]*)\")`)
	if err != nil {
		return clients, err
	}
	for _, match := range expr.FindAllString(body, -1) {
		match = strings.Replace(match, " ", "", -1)
		data := strings.Split(match, ",")
		client := Client{
			Name:  strings.Replace(data[0], "\"", "", -1),
			MAC:   strings.Replace(data[1], "\"", "", -1),
			Addr:  strings.Replace(data[2], "\"", "", -1),
			Lease: parseLease(strings.Replace(data[3], "\"", "", -1)),
		}
		clients = append(clients, client)
	}
	return clients, nil
}

// Get LAN Traffic takes as argument the list of clients obtained with GetClients
// and fill the fields related to packet and kbytes usage.
// GetClients does not returns the clients connected to the LAN, this function
// also add the ones connected through ethernet to the list.
func (r *Router) GetLANTraffic(clients []Client) ([]Client, error) {
	var enhClients []Client
	body, err := r.Get("http://" + r.Address + "/" + r.Token + STATS_URL)
	if err != nil {
		return enhClients, err
	}
	expr, err := regexp.Compile(`(?m)\d+, "([^\"]*)", "([^\"]*)", \d+, \d+`)
	if err != nil {
		return enhClients, err
	}
	for _, match := range expr.FindAllString(body, -1) {
		match = strings.Replace(match, " ", "", -1)
		data := strings.Split(match, ",")
		addr := strings.Replace(data[1], "\"", "", -1)
		mac := strings.Replace(data[2], "\"", "", -1)
		packets, err := strconv.ParseFloat(strings.Replace(data[3], "\"", "", -1), 64)
		if err != nil {
			return enhClients, err
		}
		bytes, err := strconv.ParseFloat(strings.Replace(data[4], "\"", "", -1), 64)
		if err != nil {
			return enhClients, err
		}
		found := false
		for _, client := range clients {
			if client.MAC == mac {
				client.Packets = packets
				client.Bytes = bytes / 1024
				enhClients = append(enhClients, client)
				found = true
			}
		}
		if found == false {
			client := Client{
				Name:    "ethdev",
				MAC:     mac,
				Addr:    addr,
				Lease:   0,
				Packets: packets,
				Bytes:   bytes / 1024,
			}
			enhClients = append(enhClients, client)
		}
	}
	return enhClients, nil
}

// Logout logs out of the router
func (r *Router) Logout() error {
	_, err := r.Get("http://" + r.Address + "/" + r.Token + LOGOUT_URL)
	if err != nil {
		return err
	}
	r.Token = ""
	return nil
}

// Reboot reboots the router
func (r *Router) Reboot() error {
	_, err := r.Get("http://" + r.Address + "/" + r.Token + REBOOT_URL)
	if err != nil {
		return err
	}
	r.Token = ""
	return nil
}
