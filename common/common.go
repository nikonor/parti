package common

import "gopkg.in/mgo.v2/bson"

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"strings"
	"time"
    "net"
    "errors"
)

// ObjectIDGen - получение ObjectID похожий на MongaDBID
func ObjectIDGen() string {
	u := bson.NewObjectId().Hex()
	return u
}

//MakeMD5String - получение md5 суммы от набора строк
func MakeMD5String(sep string, strs ...string) string {
	in := ""
	for _, str := range strs {
		in = in + str
	}

	md5sum := md5.Sum([]byte(in))
	hexarray := make([]string, len(md5sum))
	for i, c := range md5sum {
		hexarray[i] = hex.EncodeToString([]byte{c})
	}
	return strings.Join(hexarray, sep)

}

// MSecToTime - перевод миллисекунд в time.Time
func MSecToTime(i int64) time.Time {
	t := time.Unix(i/1000, 0)
	return t
}

// UnixTimeStamp - получение timestamp
func UnixTimeStamp(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandStringRunes - получение случайной строки
//              par:
//                      md5 - эмуляция md5
func RandStringRunes(n int, par ...string) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	if len(par) > 0 {
		if par[0] == "md5" {
			letterRunes = []rune("abcdef0123456789")
		}
	}
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func GetSelfIP() (string, error) {
    ifaces, err := net.Interfaces()
    if err != nil {
        return "", err
    }
    for _, iface := range ifaces {
        if iface.Flags&net.FlagUp == 0 {
            continue // interface down
        }
        if iface.Flags&net.FlagLoopback != 0 {
            continue // loopback interface
        }
        addrs, err := iface.Addrs()
        if err != nil {
            return "", err
        }
        for _, addr := range addrs {
            var ip net.IP
            switch v := addr.(type) {
            case *net.IPNet:
                ip = v.IP
            case *net.IPAddr:
                ip = v.IP
            }
            if ip == nil || ip.IsLoopback() {
                continue
            }
            ip = ip.To4()
            if ip == nil {
                continue // not an ipv4 address
            }
            return ip.String(), nil
        }
    }
    return "", errors.New("are you connected to the network?")
}