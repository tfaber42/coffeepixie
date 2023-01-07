# coffeepixie

1. Download Raspberry Pi Imager: https://www.raspberrypi.com/software/
2. Write Raspberry Pi OS *LITE* 32 bit image to SD card (set hostname, password, enable SSH, WLAN and locale before writing image)
3. Connect via SSH (using hostname set before writing image):
```
sudo apt-get update
sudo apt-get upgrade
sudo apt-get install git
```
4. Browse to https://golang.org/dl/ and copy the link for the latest Linux/ARM v6 version, and `wget` it, e.g.
```
cd ~
wget https://go.dev/dl/go1.19.4.linux-armv6l.tar.gz
sudo tar -C /usr/local -xzf go1.19.4.linux-armv6l.tar.gz
rm go1.19.4.linux-armv6l.tar.gz
mkdir go
nano ~/.profile

```
and append these two lines at the end:
```
PATH=$PATH:/usr/local/go/bin
GOPATH=$HOME/go
```
5. Optional if you want to access files on the RasPi: Install SMB support https://www.raspberrypi.com/documentation/computers/remote-access.html#sharing-a-folder-from-your-raspberry-pi
```
sudo apt install samba samba-common-bin smbclient cifs-utils
sudo smbpasswd -a pi
sudo nano /etc/samba/smb.conf
```
Change the `[homes]` section to make the local users' home dirs browseable and writeable:
```
[homes]
...
   browseable = yes
   read only = no
```
6. Reboot:
```
sudo reboot
```
7. Configure `git`, see https://docs.github.com/en/account-and-profile/setting-up-and-managing-your-personal-account-on-github/managing-email-preferences/setting-your-commit-email-address
```
git config --global user.name "<your username>"
git config --global user.email "<your email address>"
```
8. Install GitHub CLI: https://github.com/cli/cli/blob/trunk/docs/install_linux.md, configure (accept all defaults while being logged in to github.com), clone repo and run code:
```
gh auth login
cd ~/go/github.com/tfaber42 
gh repo clone tfaber42/coffeepixie
cd coffeepixie
go run src/main.go
```
9. Navigate to `http://<hostname>:8080` and set your coffee making time!