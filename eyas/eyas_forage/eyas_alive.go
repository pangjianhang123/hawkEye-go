package eyas_forage

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ricky1122alonefe/hawkEye-go/module"
)

func eyasAlive(masterAddr, localIp, msg string) error {
	var (
		req                 module.KeepAliveRequest
		request             *http.Request
		err                 error
		byteData, respBytes []byte
		resp                *http.Response
	)
	timeStamp := time.Now().String()
	req.Addr = localIp
	req.TimeStamp = timeStamp
	req.Msg = msg

	if byteData, err = json.Marshal(req); err != nil {
		log.Critical(err.Error())
		return err
	}

	reader := bytes.NewReader(byteData)
	if request, err = http.NewRequest("POST", masterAddr, reader); err != nil {
		log.Critical(err.Error())
		return err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	client := http.Client{}

	if resp, err = client.Do(request); err != nil {
		log.Critical(err.Error())
		return err
	}

	if respBytes, err = ioutil.ReadAll(resp.Body); err != nil {
		log.Critical(err.Error())
		return err
	}

	log.Info(string(respBytes))
	return nil
}
