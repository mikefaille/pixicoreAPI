package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/ClubCedille/pixicoreAPI/pkg/config"
	"github.com/ClubCedille/pixicoreAPI/pkg/server"
	"github.com/ClubCedille/pixicoreAPI/pkg/sshclient"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type Controller struct {
	currentConfig *config.ConfigFile
}
type DhcpdServer struct {
	Mac string
	Ip  string
}

//Getlocal pixicore demands
func (ctrl *Controller) Getlocal(c *gin.Context) {
	c.JSON(200, "success")

}

//BootServers called by pixicore client to register a new server
func (ctrl *Controller) BootServer(c *gin.Context) {

	servers, err := ctrl.currentConfig.GetServers()
	if err != nil {
		log.Warn(err)
	}

	macAddr := c.Param("macAddress")

	err = servers.AddServer(macAddr)
	if err != nil {
		log.Warn(err)
	}

	server, err := servers.GetServer(macAddr)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusNotFound, gin.H{"status": err})
	} else {

		pxeSpec := server.Boot()
		c.JSON(200, pxeSpec)
	}

}

//InstallServer Install a single server
func (ctrl *Controller) InstallServer(c *gin.Context) {
	macAddr := c.Param("macAddress")
	servers := ctrl.currentConfig.Servers
	if _, err := ctrl.currentConfig.Servers.GetServer(macAddr); err == nil {

		err := fmt.Sprint("This Requested server doesn't exist : ", macAddr)
		c.JSON(http.StatusNotFound, gin.H{"status": err})
	}

	ctrl.CollectServerInfo((*servers)[c.Param("macAddress")])

	ctrl.currentConfig.WriteYamlConfig()

	c.JSON(200, (*servers)[c.Param("macAddress")])
}

//InstallAll install all the servers available
func (ctrl *Controller) InstallAll(c *gin.Context) {
	servers := ctrl.currentConfig.Servers
	for svr := range *servers {
		ctrl.CollectServerInfo((*servers)[svr])
	}

	ctrl.currentConfig.WriteYamlConfig()

	c.JSON(200, &servers)
}

//CollectServerInfo collect information about a server with ssh
func (ctrl *Controller) CollectServerInfo(currentServer *server.Server) {

	sshConfig := ssh.ClientConfig{
		User: "core",
		Auth: []ssh.AuthMethod{
			sshclient.PublicKeyFile("~/.ssh/id_rsa"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	clientSSH := sshclient.SSHClient{
		Config: &sshConfig,
		Host:   currentServer.IPAddress,
		Port:   22,
	}

	// run command with ssh
	kernel, err := clientSSH.RunCommand("uname -r")
	if err != nil {
		log.Errorf("command run error: %s", err)
	}

	macAddressFirst, err := clientSSH.RunCommand("cat /sys/class/net/enp4s0/address")
	if err != nil {
		log.Errorf("command run error: %s\n", err)
	}
	macAddressSecond, err := clientSSH.RunCommand("cat /sys/class/net/enp5s0/address")
	if err != nil {
		log.Errorf("command run error: %s\n", err)
	}

	if currentServer.MacAddress.String() == strings.TrimSuffix(macAddressFirst, "\r\n") {
		currentServer.SecondMacAddress = strings.TrimSuffix(macAddressSecond, "\r\n")
	} else {
		currentServer.SecondMacAddress = strings.TrimSuffix(macAddressFirst, "\r\n")
	}
	currentServer.Kernel = strings.TrimSuffix(kernel, "\r\n")

	_, err = clientSSH.RunCommand("sudo coreos-install -d /dev/sda -i /run/ignition.json -C stable")
	if err != nil {
		log.Errorf("command run error: %s\n", err)
	}
	currentServer.Installed = true

}

// GetServers return config of the all the servers
func (ctrl *Controller) GetServers(c *gin.Context) {
	servers, err := ctrl.currentConfig.GetServers()

	if err != nil {
		log.Error(err)
		c.JSON(403, gin.H{"status": err})
	}

	c.JSON(200, gin.H{"success": servers})
}

func (ctrl *Controller) UpdateTest(c *gin.Context) {
	ipaddress := getServerIP(c.Param("macAddress"))
	c.JSON(200, gin.H{"reponse": ipaddress})
}

func getServerIP(macAddress string) string {

	var servers []DhcpdServer

	var httpClient = &http.Client{Timeout: 10 * time.Second}
	res, _ := httpClient.Get("http://localhost:8000")
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	json.Unmarshal(body, &servers)

	for _, element := range servers {

		if element.Mac == macAddress {
			if verifyServerConnection(element.Ip) {
				return element.Ip
			}
			return "false"
		}
	}

	return "false"
}

func verifyServerConnection(ipAddress string) bool {
	var exitCode int
	for i := 0; i < 3; i++ {

		//time.Sleep(6 * time.Second)

		cmd := exec.Command("nc -vz 142.137.247.120 22")
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}

		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()

		if exitCode == 0 {
			return true
		}
	}
	return false
}
