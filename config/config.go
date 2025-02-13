package config

import (
	"chatgpt-api-server/utility"
	"math/rand"
	"time"

	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcache"
	"github.com/gogf/gf/v2/util/gconv"
)

func CHATPROXY(ctx g.Ctx) string {
	return g.Cfg().MustGetWithEnv(ctx, "CHATPROXY").String()
}

func PLUSN(ctx g.Ctx) int {
	ab := g.Cfg().MustGetWithEnv(ctx, "PLUSN").Int()
	if ab <= 0 {
		return 1
	}
	return ab
}

func AUTHKEY(ctx g.Ctx) string {
	// g.Log().Debug(ctx, "config.AUTHKEY", g.Cfg().MustGetWithEnv(ctx, "AUTHKEY").String())
	return g.Cfg().MustGetWithEnv(ctx, "AUTHKEY").String()
}

func USERTOKENLOCK(ctx g.Ctx) bool {
	return g.Cfg().MustGetWithEnv(ctx, "USERTOKENLOCK").Bool()
}

var (
	DefaultModel     = "text-davinci-002-render-sha"
	FreeModels       = garray.NewStrArray()
	PlusModels       = garray.NewStrArray()
	NormalSet        = utility.NewSafeQueue()
	PlusSet          = utility.NewSafeQueue()
	MAXTIME          = 0
	TraceparentCache = gcache.New()
)

func PORT(ctx g.Ctx) int {
	// g.Log().Debug(ctx, "config.PORT", g.Cfg().MustGetWithEnv(ctx, "PORT").Int())
	if g.Cfg().MustGetWithEnv(ctx, "PORT").Int() == 0 {
		return 8001
	}
	return g.Cfg().MustGetWithEnv(ctx, "PORT").Int()
}

func ISFREE(ctx g.Ctx) bool {
	return g.Cfg().MustGetWithEnv(ctx, "ISFREE").Bool()
}

func APIAUTH(ctx g.Ctx) string {
	return g.Cfg().MustGetWithEnv(ctx, "APIAUTH").String()
}
func CRONINTERVAL(ctx g.Ctx) string {
	// 生成随机时间的每3天执行一次的表达式，格式为：秒 分 时 天 月 星期
	// 生成随机秒数 在0-59之间
	second := generateRandomNumber(59)
	secondStr := gconv.String(second)
	// 生成随机分钟数 在0-59之间
	minute := generateRandomNumber(59)
	minuteStr := gconv.String(minute)
	// 生成随机小时数 在0-23之间
	hour := generateRandomNumber(23)
	hourStr := gconv.String(hour)
	// 拼接cron表达式
	cronStr := secondStr + " " + minuteStr + " " + hourStr + " * * *"
	return cronStr

}

func generateRandomNumber(max int) int {
	rand.Seed(time.Now().UnixNano()) // 使用当前时间作为随机数生成器的种子
	return rand.Intn(max)            // 生成0到59之间的随机数
}

// continue
func CONTINUEMAX(ctx g.Ctx) int {
	if g.Cfg().MustGetWithEnv(ctx, "CONTINUEMAX").IsEmpty() {
		return 3
	}
	return g.Cfg().MustGetWithEnv(ctx, "CONTINUEMAX").Int()
}

type CacheSession struct {
	Email        string `json:"email"`
	AccessToken  string `json:"accessToken"`
	IsPlus       int    `json:"isPlus"`
	CooldownTime int64  `json:"cooldownTime"`
}

func init() {
	FreeModels.Append("text-davinci-002-render-sha")
	FreeModels.Append("text-davinci-002-render-sha-mobile")
	PlusModels.Append("gpt-4")
	PlusModels.Append("gpt-4-browsing")
	PlusModels.Append("gpt-4-plugins")
	PlusModels.Append("gpt-4-mobile")
	PlusModels.Append("gpt-4-dalle")
	PlusModels.Append("gpt-4-code-interpreter")
	PlusModels.Append("gpt-4-gizmo")

}
func GenerateID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// rand.Seed(time.Now().UnixNano())

	id := "chatcmpl-"
	for i := 0; i < length; i++ {
		id += string(charset[rand.Intn(len(charset))])
	}
	return id
}
