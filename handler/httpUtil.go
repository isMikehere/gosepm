package handler

import (
	"io/ioutil"
	"net/http"
	"net/url"
)

/**
get request ,response string
**/
func HttpGet(url string) (bool, string) {

	resp, err := http.Get(url)
	if err != nil {
		return false, ""
	} else {
		//一定要关闭
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		if resp.StatusCode == 200 { //相应成功
			bodyStr := string(body)
			return true, bodyStr
		} else {
			return false, ""
		}
	}
}

func HttpPost(path string, params map[string]string) (bool, string) {

	client := &http.Client{}
	reqest, _ := http.NewRequest("POST", path, nil)

	// reqest.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	reqest.Header.Set("Accept-Charset", "utf-8;q=0.7,*;q=0.3")
	// reqest.Header.Set("Accept-Language", "zh-CN,zh;q=0.8,en;q=0.6")
	reqest.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")
	//传递参数
	reqest.Form = make(url.Values)
	for k, v := range params {
		reqest.Form.Add(k, v)
	}
	resp, _ := client.PostForm(path, reqest.Form)
	// resp, _ := client.Do(reqest)
	//一定要关闭
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == 200 { //相应成功
		bodyStr := string(body)
		return false, bodyStr
	} else {
		return false, ""
	}
	return false, ""
}
