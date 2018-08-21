package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"parti/config"
	l "parti/logger"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
	"time"
)

var (
	// ErrMD5Problem - ошибка md5 конфига
	ErrMD5Problem        = errors.New("Не совпадают md5 сумму конфигов в модуле управления и в инстансе")
	ErrObligativeAdminka = errors.New("Нет связи с админкой, а она обязательная")
)

// LogoutFromAdm - отключение инстанса
func LogoutFromAdm(url string) error {
	var (
		out  []byte
		err  error
		outC Container
	)

	if url == "" {
		return nil
	}

	url = "http://" + url + "/module/logout"

	jMeta, _ := json.Marshal(Meta)

	if out, err = PostByURL(url, 1, jMeta); err != nil {
		return err
	}

	l.Debug("Logout::parti::" + string(out))

	err = json.Unmarshal(out, &outC)
	switch {
	case err != nil:
		return err
	case outC.ErrorCode != 0:
		return errors.New(outC.ErrorMessage)
	}

	return nil
}

// LoginToAdm - регистрация в адимнке
func LoginToAdm(url string, cfg config.Config) (int, error) {
	var (
		out  []byte
		err  error
		outC LoginResponse
	)

	if url == "" {
		return -1, nil
	}
	url = "http://" + url + "/module/login"

	jMeta, _ := json.Marshal(Meta)

	for tryCount := 0; tryCount < cfg.Module.Instance.Attempts; tryCount++ {
		if out, err = PostByURL(url, cfg.Module.Instance.TimeOut, jMeta); err != nil {
			return -1, err
		}

		l.Debug("Login::parti::" + string(out))

		err = json.Unmarshal(out, &outC)
		switch {
		case err != nil:
			return -1, err
		case outC.ErrorCode != 0:
			if strings.Contains(outC.ErrorMessage, "md5") {
				println(outC.ErrorCode, outC.ErrorMessage)
				l.Debug("Try Login. md5 error")
			} else {
				return -1, errors.New(outC.ErrorMessage)
			}
		case err == nil:
			// fmt.Printf("Порядковы номер=%d\n", outC.Result[0].OrderNumber)
			// fmt.Printf("\t%#v\n", outC.Result[0])
			return outC.Result[0].OrderNumber, nil
		}
		if tryCount >= 2 {
			break
		}
		time.Sleep(time.Duration(math.Pow(float64(cfg.Module.Instance.LoginTimeOut), float64(tryCount+1))) * time.Second)
	}

	return -1, nil
}

// GetByURL - отправка данных по GET
func GetByURL(url string, timeout int) ([]byte, error) {
	var (
		err  error
		ret  []byte
		resp *http.Response
	)

	cli := http.Client{Timeout: time.Duration(timeout) * time.Second}

	if resp, err = cli.Get(url); err != nil {
		return ret, err
	}
	if ret, err = ioutil.ReadAll(resp.Body); err != nil {
		return ret, err
	}
	defer resp.Body.Close()

	return ret, nil
}

// PostByURL - отправка данных по пост
func PostByURL(url string, timeout int, inJSON []byte) ([]byte, error) {
	var (
		err  error
		ret  []byte
		resp *http.Response
	)

	cli := http.Client{Timeout: time.Duration(timeout) * time.Second}

	if resp, err = cli.Post(url, "application/json", bytes.NewReader(inJSON)); err != nil {
		return ret, err
	}
	if ret, err = ioutil.ReadAll(resp.Body); err != nil {
		return ret, err
	}
	defer resp.Body.Close()

	return ret, nil
}
