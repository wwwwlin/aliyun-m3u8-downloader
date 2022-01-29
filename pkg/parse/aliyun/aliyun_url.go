package aliyun

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"github.com/bitly/go-simplejson"
	"github.com/google/uuid"
	"log"
	"net/url"
	"sort"
	"strings"
	"time"
)

var (
	PlayAuthSign1 = []int{52, 58, 53, 121, 116, 102}
	PlayAuthSign2 = []int{90, 91}
)

func GetPlayInfoRequestUrl(rand, playAuth, videoId, formats string) (string, error) {
	playAuth = decodePlayAuth(playAuth)
	sj, err := simplejson.NewJson([]byte(playAuth))
	if err != nil {
		log.Println(err)
		return "", err
	}
	// 公共参数
	publicParams := map[string]string{}
	publicParams["AccessKeyId"], _ = sj.Get("AccessKeyId").String()
	publicParams["Timestamp"] = generateTimestamp()
	publicParams["SignatureMethod"] = "HMAC-SHA1"
	publicParams["SignatureVersion"] = "1.0"
	publicParams["SignatureNonce"] = uuid.NewString()
	publicParams["Format"] = "JSON"
	publicParams["Channel"] = "HTML5"
	publicParams["StreamType"] = "video"
	if len(rand) > 0 {
		publicParams["Rand"] = rand
	}
	publicParams["Formats"] = ""
	if len(formats) > 0 {
		publicParams["Formats"] = formats
	}
	publicParams["Version"] = "2017-03-21"
	// 私有参数
	privateParams := map[string]string{}
	privateParams["Action"] = "GetPlayInfo"
	privateParams["AuthInfo"], _ = sj.Get("AuthInfo").String()
	privateParams["AuthTimeout"] = "7200"
	privateParams["Definition"] = "240" //gk 此参数为空
	privateParams["PlayConfig"] = "{}"
	privateParams["PlayerVersion"] = "2.8.2"
	privateParams["ReAuthInfo"] = "{}"
	privateParams["SecurityToken"], _ = sj.Get("SecurityToken").String()
	privateParams["VideoId"] = videoId
	allParams := getAllParams(publicParams, privateParams)
	cqs := getCQS(allParams)
	stringToSign := "GET" + "&" + percentEncode("/") + "&" + percentEncode(cqs)
	accessKeySecret, _ := sj.Get("AccessKeySecret").String()
	signature := hmacSHA1Signature(accessKeySecret, stringToSign)
	// query
	queryString := cqs + "&Signature=" + percentEncode(signature)
	return "https://vod.cn-shanghai.aliyuncs.com/?" + queryString, nil
}

func hmacSHA1Signature(accessKeySecret, stringToSign string) string {
	key := accessKeySecret + "&"
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func percentEncode(s string) string {
	return url.QueryEscape(s)
}

func getCQS(allParams []string) string {
	sort.Strings(allParams)
	return strings.Join(allParams, "&")
}

func getAllParams(publicParams, privateParams map[string]string) (allParams []string) {
	for key, value := range publicParams {
		allParams = append(allParams, percentEncode(key)+"="+percentEncode(value))
	}
	for key, value := range privateParams {
		allParams = append(allParams, percentEncode(key)+"="+percentEncode(value))
	}
	return allParams
}

func decodePlayAuth(playAuth string) string {
	if isSignedPlayAuth(playAuth) {
		playAuth = decodeSignedPlayAuth2B64(playAuth)
	}
	data, err := base64.StdEncoding.DecodeString(playAuth)
	if err != nil {
		return ""
	}
	return string(data)
}

func isSignedPlayAuth(playAuth string) bool {
	signPos1 := time.Now().Year() / 100 // 当前年份
	signPos2 := len(playAuth) - 2
	sign1 := getSignStr(PlayAuthSign1)
	sign2 := getSignStr(PlayAuthSign2)
	r1 := playAuth[signPos1 : signPos1+len(sign1)]
	r2 := playAuth[signPos2:]
	return sign1 == r1 && r2 == sign2
}

func decodeSignedPlayAuth2B64(playAuth string) string {
	sign1 := getSignStr(PlayAuthSign1)
	sign2 := getSignStr(PlayAuthSign2)
	playAuth = strings.Replace(playAuth, sign1, "", 1)
	playAuth = playAuth[:len(playAuth)-len(sign2)]
	factor := time.Now().Year() / 100 // 当前年份
	newCharCodeList := []byte(playAuth)
	for i, code := range newCharCodeList {
		r := int(code) / factor
		z := factor / 10
		if r == z {
			newCharCodeList[i] = code
		} else {
			newCharCodeList[i] = code - 1
		}
	}
	return string(newCharCodeList)
}

func getSignStr(sign []int) string {
	s := strings.Builder{}
	for i, b := range sign {
		s.WriteByte(byte(b - i))
	}
	return s.String()
}

func generateTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
