package utils

import (
	"errors"
	"fmt"
	"hash/fnv"
	"math/rand"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unsafe"

	"github.com/bytedance/sonic"
	"golang.org/x/exp/constraints"
)

func RunFuncName() string {
	const index = 2
	pc := make([]uintptr, 1)
	runtime.Callers(index, pc)
	f := runtime.FuncForPC(pc[0])
	sl := strings.Split(f.Name(), "/")
	return sl[len(sl)-1]
}

const MaxPrintBodyLen = 10240 // 10KB

func LimitedLogBody(body []byte) []byte {
	if len(body) > MaxPrintBodyLen {
		dst := make([]byte, MaxPrintBodyLen)
		copy(dst, body)
		return dst
	}
	return body
}

func IsExistFile(p string) (bool, error) {
	if _, err := os.Stat(p); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, nil
	}
}

func TimeString(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func DateString(t time.Time) string {
	return t.Format("20060102")
}

func TimeNow() string {
	return TimeString(time.Now())
}

func TimestampNow() int {
	return int(time.Now().Unix())
}

func DateNow() string {
	return DateString(time.Now())
}

func MinutesToNextHour() time.Duration {
	return time.Duration(60-time.Now().Minute()) * time.Minute
}

func MustMarshal(d interface{}) string {
	if d == nil {
		return ""
	}

	if d, err := sonic.Marshal(d); err != nil {
		return ""
	} else {
		return string(d)
	}
}

func MustMarshalBytes(d interface{}) []byte {
	if d, err := sonic.Marshal(d); err != nil {
		return nil
	} else {
		return d
	}
}

func MustMarshalIndent(d interface{}) string {
	if d, err := sonic.ConfigDefault.MarshalIndent(d, "", "\t"); err != nil {
		return ""
	} else {
		return string(d)
	}
}

func MustInt(d string) int {
	res, err := strconv.Atoi(d)
	if err != nil {
		return 0
	}
	return res
}

func MustWithDefault[T constraints.Integer | constraints.Float](raw string, defaultValue T) T {
	var val T
	if _, err := fmt.Sscan(raw, &val); err != nil {
		return defaultValue
	}
	return val
}

func MustFloat32(d string) float32 {
	res, err := strconv.ParseFloat(d, 32)
	if err != nil {
		return 0
	}
	return float32(res)
}

func StringToImmutableBytes(s string) []byte {
	var b []byte
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	bh.Data = (*reflect.StringHeader)(unsafe.Pointer(&s)).Data
	bh.Len = len(s)
	bh.Cap = len(s)
	return b
}

func ImmutableBytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func GetEnvWithDefault(key, defaultV string) string {
	if v, exist := os.LookupEnv(key); exist {
		return v
	} else {
		return defaultV
	}
}

func GetEnv(key string) string {
	return GetEnvWithDefault(key, "")
}

// GetLocalIP returns the non loopback local IP of the host
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func GetHostName() string {
	if res, err := os.Hostname(); err != nil {
		return res
	}
	return ""
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func StringSliceContains(sl []string, ss string) bool {
	if len(sl) == 0 {
		return false
	}

	for _, s := range sl {
		if s == ss {
			return true
		}
	}
	return false
}

func MustDuration(d string) time.Duration {
	if dd, err := time.ParseDuration(d); err != nil {
		return 0
	} else {
		return dd
	}
}

type VkeInfo struct {
	PodName  string `json:"pod_name"`
	PodIp    string `json:"pod_ip"`
	NodeName string `json:"node_name"`
	HostIp   string `json:"host_ip"`
}

func GetVkeInfoString() string {
	return fmt.Sprintf("{pod_name=%s: pod_ip=%s: node_name=%s: host_ip=%s}",
		os.Getenv("VKE_POD_NAME"),
		os.Getenv("VKE_POD_IP"),
		os.Getenv("VKE_NODE_NAME"),
		os.Getenv("VKE_HOST_IP"),
	)
}

func NewSingleErrChannel(err error) <-chan error {
	res := make(chan error, 1)
	res <- err
	close(res)
	return res
}

func GetErrFromErrChannel(ec <-chan error) error {
	if ec == nil {
		return nil
	}

	select {
	case e, ok := <-ec:
		// Note: actually not need to check ok: if channel is closed(ok=false), will return nil error
		if ok {
			return e
		} else {
			return nil
		}
	default:
		return nil
	}
}

func StringGetDefault(str, defaultStr string) string {
	if str == "" {
		return defaultStr
	} else {
		return str
	}
}

func PtrValueOrError[T any](v any, err error) (*T, error) {
	if err != nil {
		return nil, err
	}
	return v.(*T), nil
}

func ValueOrError[T any](v any, err error) (T, error) {
	if err != nil {
		var empty T
		return empty, err
	}
	return v.(T), nil
}

func Max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Min[T constraints.Ordered](a, b T) T {
	if a > b {
		return b
	}
	return a
}

func Hash(key string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return h.Sum32()
}

func Sample(n int) bool {
	if rand.Intn(100) < n {
		return true
	}
	return false
}

func ExtractBucketAndKey(sourcePath string) (string, string, string, error) { // from MLP
	u, err := url.Parse(sourcePath)
	if err != nil {
		return "", "", "", err
	}
	// get bucket
	bucket := u.Host
	key := strings.TrimLeft(u.Path, "/")
	return bucket, key, u.Scheme, nil
}

func GetFileExtension(filename string) string {
	return filepath.Ext(filename)
}

func ChineseCount(str string) (count int) {
	count = 0
	for _, char := range str {
		if unicode.Is(unicode.Han, char) {
			count++
		}
	}
	return
}

func Format(str string) (int, error) {
	multi := 1
	if strings.HasSuffix(str, "K") || strings.HasSuffix(str, "k") {
		multi = 1000
		str = strings.Replace(strings.ToLower(str), "k", "", 1)
	} else if strings.HasSuffix(str, "M") || strings.HasSuffix(str, "m") {
		multi = 1000 * 1000
		str = strings.Replace(strings.ToLower(str), "m", "", 1)
	}
	num, err := strconv.Atoi(str)
	if err != nil {
		return num, err
	}
	return num * multi, nil
}

func P[T any](v T) *T {
	return &v
}

// SliceInsertElement insert a new element into specified position of a slice and return a new slice
func SliceInsertElement[T any](slice []T, index int, value T) []T {
	newSlice := make([]T, len(slice)+1)
	copy(newSlice[:index], slice[:index])
	newSlice[index] = value
	copy(newSlice[index+1:], slice[index:])

	return newSlice
}

func MustString[T ~string](p *T) T {
	if p == nil {
		return ""
	}
	return *p
}

func BoolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func All[T any](slice []T, condition func(T) bool) bool {
	for _, t := range slice {
		if !condition(t) {
			return false
		}
	}
	return true
}

// GetPages 计算页数
func GetPages(pageSize, count int) (pages int) {
	//pages = int(math.Ceil(float64(raw.Data.TotalHits) / float64(EastmoneyNoticesPageSize)))
	pages = count / pageSize
	n := count % pageSize
	if n > 0 {
		pages++
	}
	return pages
}
