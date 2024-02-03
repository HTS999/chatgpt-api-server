package chat

import (
	"chatgpt-api-server/apireq"
	"chatgpt-api-server/apirespstream"
	backendapi "chatgpt-api-server/backend-api"
	"chatgpt-api-server/config"
	"chatgpt-api-server/modules/chatgpt/model"
	"chatgpt-api-server/utility"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/cool-team-official/cool-admin-go/cool"
	"github.com/gogf/gf/text/gregex"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/launchdarkly/eventsource"
)

var (
	// client    = g.Client()
	ErrNoAuth = `{
		"error": {
			"message": "You didn't provide an API key. You need to provide your API key in an Authorization header using Bearer auth (i.e. Authorization: Bearer YOUR_KEY), or as the password field (with blank username) if you're accessing the API from your browser and are prompted for a username and password. You can obtain an API key from https://platform.openai.com/account/api-keys.",
			"type": "invalid_request_error",
			"param": null,
			"code": null
		}
	}`
	ErrKeyInvalid = `{
		"error": {
			"message": "Incorrect API key provided: sk-4yNZz***************************************6mjw. You can find your API key at https://platform.openai.com/account/api-keys.",
			"type": "invalid_request_error",
			"param": null,
			"code": "invalid_api_key"
		}
	}`
	ChatReqStr = `{
		"action": "next",
		"messages": [
		  {
			"id": "aaa2f210-64e1-4f0d-aa51-e73fe1ae74af",
			"author": { "role": "user" },
			"content": { "content_type": "text", "parts": ["1\n"] },
			"metadata": {}
		  }
		],
		"parent_message_id": "aaa1a8ab-61d6-4fc0-a5f5-181015c2ebaf",
		"model": "text-davinci-002-render-sha",
		"timezone_offset_min": -480,
		"suggestions": [],
		"history_and_training_disabled": true,
		"conversation_mode": { "kind": "primary_assistant" }
	  }
	  `
	ChatTurboReqStr = `
	{
		"action": "next",
		"messages": [
			{
				"id": "aaa2b2cc-e7e9-47c5-8171-0ff8a6d9d6d3",
				"author": {
					"role": "user"
				},
				"content": {
					"content_type": "text",
					"parts": [
						"你好"
					]
				},
				"metadata": {}
			}
		],
		"parent_message_id": "aaa1403d-c61e-4818-90e0-93a99465aec6",
		"model": "gpt-4",
		"timezone_offset_min": -480,
		"suggestions": [],
		"history_and_training_disabled": true,
		"conversation_mode": {
			"kind": "primary_assistant"
		}
	}`
	Chat4ReqGizmo = `
	{
		"action": "next",
		"messages": [
			{
				"id": "aaa29c6e-0b9a-4992-acdf-14658caf65ab",
				"author": {
					"role": "user"
				},
				"content": {
					"content_type": "text",
					"parts": [
						"你是谁？"
					]
				},
				"metadata": {}
			}
		],
		"parent_message_id": "aaa1fc87-cdd0-41dc-92df-3230c63b4d7a",
		"model": "gpt-4-gizmo",
		"timezone_offset_min": -480,
		"suggestions": [],
		"history_and_training_disabled": false,
		"conversation_mode": {
			"kind": "gizmo_interaction",
			"gizmo_id": "g-YyyyMT9XH"
		},
		"force_paragen": false,
		"force_rate_limit": false
	}
	`
	Chat4ReqStr = `
	{
		"action": "next",
		"messages": [
			{
				"id": "aaa2b182-d834-4f30-91f3-f791fa953204",
				"author": {
					"role": "user"
				},
				"content": {
					"content_type": "text",
					"parts": [
						"画一只猫1231231231"
					]
				},
				"metadata": {}
			}
		],
		"parent_message_id": "aaa11581-bceb-46c5-bc76-cb84be69725e",
		"model": "gpt-4-gizmo",
		"timezone_offset_min": -480,
		"suggestions": [],
		"history_and_training_disabled": true,
		"conversation_mode": {
			"gizmo": {
				"gizmo": {
					"id": "g-YyyyMT9XH",
					"organization_id": "org-OROoM5KiDq6bcfid37dQx4z4",
					"short_url": "g-YyyyMT9XH-chatgpt-classic",
					"author": {
						"user_id": "user-u7SVk5APwT622QC7DPe41GHJ",
						"display_name": "ChatGPT",
						"selected_display": "name",
						"is_verified": true
					},
					"voice": {
						"id": "ember"
					},
					"display": {
						"name": "ChatGPT Classic",
						"description": "The latest version of GPT-4 with no additional capabilities",
						"welcome_message": "Hello",
						"profile_picture_url": "https://files.oaiusercontent.com/file-i9IUxiJyRubSIOooY5XyfcmP?se=2123-10-13T01%3A11%3A31Z&sp=r&sv=2021-08-06&sr=b&rscc=max-age%3D31536000%2C%20immutable&rscd=attachment%3B%20filename%3Dgpt-4.jpg&sig=ZZP%2B7IWlgVpHrIdhD1C9wZqIvEPkTLfMIjx4PFezhfE%3D",
						"categories": []
					},
					"share_recipient": "link",
					"updated_at": "2023-11-06T01:11:32.191060+00:00",
					"last_interacted_at": "2023-11-18T07:50:19.340421+00:00",
					"tags": [
						"public",
						"first_party"
					]
				},
				"tools": [],
				"files": [],
				"product_features": {
					"attachments": {
						"type": "retrieval",
						"accepted_mime_types": [
							"text/x-c",
							"text/html",
							"application/x-latext",
							"text/plain",
							"text/x-ruby",
							"text/x-typescript",
							"text/x-c++",
							"text/x-java",
							"text/x-sh",
							"application/vnd.openxmlformats-officedocument.presentationml.presentation",
							"text/x-script.python",
							"text/javascript",
							"text/x-tex",
							"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
							"application/msword",
							"application/pdf",
							"text/x-php",
							"text/markdown",
							"application/json",
							"text/x-csharp"
						],
						"image_mime_types": [
							"image/jpeg",
							"image/png",
							"image/gif",
							"image/webp"
						],
						"can_accept_all_mime_types": true
					}
				}
			},
			"kind": "gizmo_interaction",
			"gizmo_id": "g-YyyyMT9XH"
		},
		"force_paragen": false,
		"force_rate_limit": false
	}`
	ApiRespStr = `{
		"id": "chatcmpl-LLKfuOEHqVW2AtHks7wAekyrnPAoj",
		"object": "chat.completion",
		"created": 1689864805,
		"model": "gpt-3.5-turbo",
		"usage": {
			"prompt_tokens": 0,
			"completion_tokens": 0,
			"total_tokens": 0
		},
		"choices": [
			{
				"message": {
					"role": "assistant",
					"content": "Hello! How can I assist you today?"
				},
				"finish_reason": "stop",
				"index": 0
			}
		]
	}`
	ApiRespStrStream = `{
		"id": "chatcmpl-afUFyvbTa7259yNeDqaHRBQxH2PLH",
		"object": "chat.completion.chunk",
		"created": 1689867370,
		"model": "gpt-3.5-turbo",
		"choices": [
			{
				"delta": {
					"content": "Hello"
				},
				"index": 0,
				"finish_reason": null
			}
		]
	}`
	ApiRespStrStreamEnd = `{"id":"apirespid","object":"chat.completion.chunk","created":apicreated,"model": "apirespmodel","choices":[{"delta": {},"index": 0,"finish_reason": "stop"}]}`
)

func Completions(r *ghttp.Request) {

	corsOptions := r.Response.DefaultCORSOptions()
	corsOptions.AllowDomain = []string{"*"}
	corsOptions.AllowHeaders = "*"
	corsOptions.AllowCredentials = "true"
	corsOptions.AllowMethods = "*"
	corsOptions.MaxAge = 3601
	r.Response.CORS(corsOptions)

	ctx := r.Context()
	// g.Log().Debug(ctx, "Conversation start....................")
	if config.MAXTIME > 0 {
		// 创建带有超时的context
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(config.MAXTIME)*time.Second)
		defer cancel()
	}

	// 获取 Header 中的 Authorization	去除 Bearer
	userToken := strings.TrimPrefix(r.Header.Get("authorization"), "Bearer ")
	// 如果 Authorization 为空，返回 401
	if userToken == "" {
		r.Response.Status = 401
		// r.Response.WriteJson(g.Map{
		// 	"detail": "Authentication credentials were not provided.",
		// })
		r.Response.WriteJson(gApiError("somethine wrong with api-key", gerror.New("openai_hk_error")))
	}
	isPlusUser := false
	if !config.ISFREE(ctx) {
		userRecord, err := cool.DBM(model.NewChatgptUser()).Where("userToken", userToken).Where("expireTime>now()").Cache(gdb.CacheOption{
			Duration: 10 * time.Minute,
			Name:     "userToken:" + userToken,
			Force:    true,
		}).One()
		if err != nil {
			g.Log().Error(ctx, err)
			r.Response.Status = 500
			// r.Response.WriteJson(g.Map{
			// 	"detail": err.Error(),
			// })
			r.Response.WriteJson(gApiError("somethine wrong with api-key", err))
			return
		}
		if userRecord.IsEmpty() {
			g.Log().Error(ctx, "userToken not found")
			r.Response.Status = 401
			// r.Response.WriteJson(g.Map{
			// 	"detail": "userToken not found",
			// })
			r.Response.WriteJson(gApiError("somethine wrong with api-key", gerror.New("openai_hk_error")))
			return
		}
		if userRecord["isPlus"].Int() == 1 {
			isPlusUser = true
		}
	}

	// g.Log().Debug(ctx, "token: ", token)
	// 从请求中获取参数
	req := &apireq.Req{}
	err := r.GetRequestStruct(req)
	if err != nil {
		g.Log().Error(ctx, "r.GetRequestStruct(req) error: ", err)
		r.Response.Status = 400
		r.Response.WriteJson(gjson.New(`{"error": "bad request"}`))
		return
	}
	// g.Dump(req)
	// 遍历 req.Messages 拼接 newMessages
	newMessages := ""
	for _, message := range req.Messages {
		newMessages += message.Content + "\n"
	}
	// g.Dump(newMessages)
	//g.Log().Info(ctx, "model", req.Model)
	var ChatReq *gjson.Json
	if gstr.HasPrefix(req.Model, "gpt-4") {
		ChatReq = gjson.New(Chat4ReqStr)
	} else {
		ChatReq = gjson.New(ChatReqStr)
	}

	if gstr.HasPrefix(req.Model, "gpt-4-all") {
		//if gstr.HasPrefix(req.Model, "gpt-4-turbo") {
		ChatReq = gjson.New(ChatTurboReqStr)

	}
	if gstr.HasPrefix(req.Model, "gpt-4-gizmo-") {
		ChatReq = gjson.New(Chat4ReqGizmo)
		//ChatReq.Set("conversation_mode.gizmo.gizmo.id", gstr.Replace(req.Model, "gpt-4-gizmo-", ""))
		ChatReq.Set("conversation_mode.gizmo_id", gstr.Replace(req.Model, "gpt-4-gizmo-", ""))
		g.Log().Info(ctx, "gizmo", gstr.Replace(req.Model, "gpt-4-gizmo-", ""))
	}
	//

	if req.Model == "gpt-4-mobile" {
		ChatReq = gjson.New(ChatReqStr)
		ChatReq.Set("model", "gpt-4-mobile")
	}

	// 如果不是plus用户但是使用了plus模型
	if !isPlusUser && gstr.HasPrefix(req.Model, "gpt-4") {
		r.Response.Status = 501
		// r.Response.WriteJson(g.Map{
		// 	"detail": "plus user only",
		// })
		r.Response.WriteJson(gApiError("You can use only model gpt-3.5", gerror.New("openai_hk_error")))
		return
	}
	email := ""
	emailPop2 := ""
	clears_in := 0
	icount := 0
	// plus失效
	isPlusInvalid := false
	// 是否归还
	isReturn := true
	isPlusModel := gstr.HasPrefix(req.Model, "gpt-4")
	if isPlusModel {
		Traceparent := r.Header.Get("Traceparent")
		// Traceparent like 00-d8c66cc094b38d1796381c255542f971-09988d8458a2352c-01 获取第二个参数
		// 以-分割，取第二个参数
		TraceparentArr := gstr.Split(Traceparent, "-")
		if len(TraceparentArr) >= 2 {
			// 获取第二个参数
			Traceparent = TraceparentArr[1]
			g.Log().Info(ctx, "Traceparent", Traceparent)
			email = config.TraceparentCache.MustGet(ctx, Traceparent).String()
		}

		defer func() {
			go func() {
				if email != "" && isReturn {
					if isPlusInvalid {
						// 如果plus失效，将isPlus设置为0
						cool.DBM(model.NewChatgptSession()).Where("email=?", email).Update(g.Map{
							"isPlus": 0,
						})
						// 从set中删除
						//config.PlusSet.Remove(email)
						config.PlusSet.Remove(emailPop2)
						// 添加到set
						config.NormalSet.Add(email)
						return
					}
					if clears_in > 0 {
						// 延迟归还
						time.Sleep(time.Duration(clears_in) * time.Second)
					}
					//config.PlusSet.Add(email)
					config.PlusSet.Add(emailPop2)
					g.Log().Info(ctx, "归还 ", email, emailPop2)

				}
			}()
		}()
		if email == "" {
			emailPop, ok := config.PlusSet.Pop()
			//g.Log().Info(ctx, emailPop, ok)
			//g.Dump(config.PlusSet)
			if !ok {
				g.Dump(config.PlusSet)
				g.Log().Error(ctx, "Get email from set error")
				r.Response.Status = 500
				// r.Response.WriteJson(g.Map{
				// 	"detail": "Server is busy, please try again later",
				// })
				r.Response.WriteJson(gApiError("Server is busy, please try again later", gerror.New("openai_hk_error")))
				return
			}

			email = emailPop
		}
		emailPop2 = email
		patternStr := `\[(\d+)\]`
		email, err = gregex.ReplaceString(patternStr, "", email)
		if err != nil {
			return
		}

	} else {
		emailPop, ok := config.NormalSet.Pop()
		if !ok {
			g.Log().Error(ctx, "Get email from set error")
			r.Response.Status = 500
			// r.Response.WriteJson(g.Map{
			// 	"detail": "Server is busy, please try again later",
			// })
			r.Response.WriteJson(gApiError("Server is busy, please try again later", gerror.New("openai_hk_error")))
			return
		}
		defer func() {
			go func() {
				if email != "" && isReturn {
					config.NormalSet.Add(email)
				}
			}()
		}()

		email = emailPop
		emailPop2 = emailPop
	}
	if email == "" {
		g.Log().Error(ctx, "Get email from set error")
		r.Response.Status = 500
		// r.Response.WriteJson(g.Map{
		// 	"detail": "Server is busy, please try again later",
		// })
		r.Response.WriteJson(gApiError("Server is busy, please try again later", gerror.New("openai_hk_error")))
		return
	}
	realModel := ChatReq.Get("model").String()
	//g.Log().Info(ctx, userToken, "使用", email, req.Model, "->", realModel, "发起会话")
	if isPlusModel {
		g.Log().Info(ctx, email, "发起", config.PlusSet.Size(), req.Model, "->", realModel)
	} else {
		g.Log().Info(ctx, email, "发起", config.NormalSet.Size(), req.Model, "->", realModel)
	}

	// 使用email获取 accessToken
	sessionCache := &config.CacheSession{}
	cool.CacheManager.MustGet(ctx, "session:"+email).Scan(&sessionCache)
	accessToken := sessionCache.AccessToken
	err = utility.CheckAccessToken(accessToken)
	if err != nil { // accessToken失效
		g.Log().Error(ctx, err)
		isReturn = false
		cool.DBM(model.NewChatgptSession()).Where("email", email).Update(g.Map{"status": 0})

		go backendapi.RefreshSession(email)
		r.Response.Status = 401
		// r.Response.WriteJson(g.Map{
		// 	"detail": "accessToken is invalid,will be refresh",
		// })
		r.Response.WriteJsonExit(gApiError("accessToken is invalid,will be refresh", gerror.New("openai_hk_error")))
		return
	}
	ChatReq.Set("messages.0.content.parts.0", newMessages)
	if !gfile.Exists("./temp") {
		err := gfile.Mkdir("./temp")
		if err != nil {
			r.Response.Status = 400
			r.Response.WriteJsonExit(gApiError("create temp dir failed ", gerror.New("openai_hk_error")))
		}
	}
	//文件下载+上传
	if gstr.Count(newMessages, "http") > 0 && (req.Model == "gpt-4-all") {
		ChatReq, newMessages, err = gpt4all(ctx, newMessages, accessToken)
		if err != nil {
			r.Response.Status = 428
			r.Response.WriteJsonExit(gApiError("file download error", err))
		}
	} else if gstr.Count(newMessages, "http") > 0 && gstr.HasPrefix(req.Model, "gpt-4-gizmo-") {

		ChatReq, err = gpt4miz(ctx, newMessages, accessToken, ChatReq)
		if err != nil {
			r.Response.Status = 428
			r.Response.WriteJsonExit(gApiError("file download error", err))
		}

	}

	ChatReq.Set("messages.0.id", uuid.NewString())
	ChatReq.Set("parent_message_id", uuid.NewString())

	if len(req.PluginIds) > 0 {
		ChatReq.Set("plugin_ids", req.PluginIds)
	}
	ChatReq.Set("history_and_training_disabled", true)

	//最后的调试
	//ChatReq.Dump()
	// 请求openai
	resp, err := g.Client().SetHeaderMap(g.MapStrStr{
		"Authorization": "Bearer " + accessToken,
		"Content-Type":  "application/json",
		"authkey":       config.AUTHKEY(ctx),
	}).Post(ctx, config.CHATPROXY(ctx)+"/backend-api/conversation", ChatReq.MustToJson())
	if err != nil {
		g.Log().Error(ctx, "g.Client().Post error: ", err)
		r.Response.Status = 500
		//r.Response.WriteJson(gjson.New(`{"detail": "internal server error"}`))
		r.Response.WriteJson(gApiError("internal server error", gerror.New("openai_hk_error")))
		return
	}
	defer resp.Close()
	if resp.StatusCode == 401 {
		g.Log().Error(ctx, "token过期,需要重新获取token", email, resp.ReadAllString())
		isReturn = false
		cool.DBM(model.NewChatgptSession()).Where("email", email).Update(g.Map{"status": 0})
		go backendapi.RefreshSession(email)
		r.Response.WriteStatus(401, resp.ReadAllString())
		return
	}
	if resp.StatusCode == 429 {
		resStr := resp.ReadAllString()

		clears_in = gjson.New(resStr).Get("detail.clears_in").Int()

		//延迟规划
		if clears_in > 0 {
			g.Log().Error(ctx, email, "resp.StatusCode==430", clears_in, resStr)

			//r.Response.WriteStatusExit(430, resStr)
			//r.Response.WriteStatusExit(430, `{ "detail": "try again" ,"StatusCode":430 }`)
			r.Response.Status = 430
			r.Response.WriteJsonExit(gApiError("try again", gerror.New("openai_hk_error")))
			return
		} else {
			g.Log().Error(ctx, email, "resp.StatusCode==429", resStr)
			r.Response.Status = 429
			r.Response.WriteJsonExit(gApiError("try again", gerror.New("openai_hk_error")))
			return
		}
	}
	// 如果返回结果不是200
	if resp.StatusCode != 200 {
		allString := resp.ReadAllString()
		g.Log().Error(ctx, "resp.StatusCode: ", resp.StatusCode, gstr.SubStr(gstr.Replace(allString, "\n", "="), 0, 100))
		r.Response.Status = resp.StatusCode
		//aJson := gjson.New(allString)
		msg := "请求发生错误！ 请重试"
		if resp.StatusCode == 404 {
			msg = "未找到这个模型"
		}
		if resp.StatusCode == 413 {
			msg = "内容过长，请新建一个聊天"
		}
		//aJson.Set("error", msg)
		//aJson.Set("StatusCode", resp.StatusCode)
		// if gstr.HasPrefix(aJson.Get("detail").String(), "help") {
		// 	aJson.Set("detail", "")

		// }
		r.Response.WriteJsonExit(gApiError(msg, gerror.New("openai_hk_error")))
		//r.Response.WriteJson(gjson.New(allString))
		return
	}
	// if resp.Header.Get("Content-Type") != "text/event-stream; charset=utf-8" && resp.Header.Get("Content-Type") != "text/event-stream" {
	// 	g.Log().Error(ctx, "resp.Header.Get(Content-Type): ", resp.Header.Get("Content-Type"))
	// 	r.Response.Status = 500
	// 	r.Response.WriteJson(gjson.New(`{"detail": "internal server error"}`))
	// 	return
	// }
	// 流式返回
	if req.Stream {
		r.Response.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		r.Response.Header().Set("Cache-Control", "no-cache")
		r.Response.Header().Set("Connection", "keep-alive")
		modelSlug := ""
		codeMsg := ""
		message := ""
		decoder := eventsource.NewDecoder(resp.Body)

		defer decoder.Decode()

		id := config.GenerateID(29)
		for {
			event, err := decoder.Decode()
			if err != nil {
				// if err == io.EOF {
				// 	break
				// }

				if icount == 0 {
					//r.Response.WriteJson(gApiError("错误 请重试", gerror.New("try again!")))
					g.Log().Info(ctx, "释放资源", icount, email)
				}
				break
			}
			text := event.Data()
			// g.Log().Debug(ctx, "text: ", text)
			if text == "" {
				continue
			}
			if text == "[DONE]" {
				apiRespStrEnd := gstr.Replace(ApiRespStrStreamEnd, "apirespid", id)
				apiRespStrEnd = gstr.Replace(apiRespStrEnd, "apicreated", gconv.String(time.Now().Unix()))
				apiRespStrEnd = gstr.Replace(apiRespStrEnd, "apirespmodel", req.Model)
				r.Response.Writefln("data: " + apiRespStrEnd + "\n\n")
				r.Response.Flush()
				r.Response.Writefln("data: " + text + "\n\n")
				r.Response.Flush()
				continue
				// resp.Close()

				// break
			}
			// gjson.New(text).Dump()
			role := gjson.New(text).Get("message.author.role").String()
			if role == "tool" {
				//code 结束的地方
				if codeMsg != "" {
					icount++
					toStream(ctx, id, "\n```\n", req, r)
					codeMsg = ""
				}
				toolStr := myFileTool(ctx, text, accessToken)
				if toolStr != "" {
					icount++
					toStream(ctx, id, "\n"+toolStr+"\n", req, r)
				}
			}
			if role == "assistant" {
				messageTemp := gjson.New(text).Get("message.content.parts.0").String()
				model_slug := gjson.New(text).Get("message.metadata.model_slug").String()
				if model_slug != "" {
					modelSlug = model_slug
				}

				if messageTemp == "" {
					g.Log().Debug(ctx, "mygod: ", text)
					//codeMsg
					msgCode2 := gjson.New(text).Get("message.content.text").String()
					if msgCode2 != "" {
						//开始的的第三
						if codeMsg == "" {
							//
							language := gjson.New(text).Get("message.content.language").String()
							if language == "unknown" {
								language = ""
							}
							icount++
							toStream(ctx, id, "\n```"+language+"\n", req, r)
						}
						content := strings.Replace(msgCode2, codeMsg, "", 1)
						toStream(ctx, id, content, req, r)
						codeMsg = msgCode2
						continue
					}
				}
				//code 结束的地方
				if codeMsg != "" {
					icount++
					toStream(ctx, id, "\n```\n", req, r)
					codeMsg = ""
				}

				// g.Log().Debug(ctx, "messageTemp: ", messageTemp)
				// 如果 messageTemp 不包含 message 且plugin_ids为空
				if !gstr.Contains(messageTemp, message) && len(req.PluginIds) == 0 {
					continue
				}

				content := strings.Replace(messageTemp, message, "", 1)
				if content == "" {
					continue
				}

				message = messageTemp
				apiResp := gjson.New(ApiRespStrStream)
				apiResp.Set("id", id)
				apiResp.Set("created", time.Now().Unix())
				apiResp.Set("choices.0.delta.content", content)
				// if req.Model == "gpt-4" {
				// 	apiResp.Set("model", "gpt-4")
				// }
				apiResp.Set("model", req.Model)
				apiRespStruct := &apirespstream.ApiRespStreamStruct{}
				gconv.Struct(apiResp, apiRespStruct)
				// g.Dump(apiRespStruct)
				// 创建一个jsoniter的Encoder
				json := jsoniter.ConfigCompatibleWithStandardLibrary

				// 将结构体转换为JSON文本并保持顺序
				sortJson, err := json.Marshal(apiRespStruct)
				if err != nil {
					fmt.Println("转换JSON出错:", err)
					continue
				}
				icount++
				r.Response.Writeln("data: " + string(sortJson) + "\n\n")
				r.Response.Flush()
			}

		}

		if realModel != "text-davinci-002-render-sha" && modelSlug == "text-davinci-002-render-sha" {
			isPlusInvalid = true
			g.Log().Info(ctx, userToken, "使用", email, realModel, "->", modelSlug, "PLUS失效")
		} else {
			//g.Log().Info(ctx, userToken, "使用", email, realModel, "->", modelSlug, "完成会话")
			g.Log().Info(ctx, email, "完成会话", icount, realModel, "->", modelSlug)
		}

	} else {
		// 非流式回应
		modelSlug := ""
		content := ""
		msgCode := ""
		msgTool := ""
		decoder := eventsource.NewDecoder(resp.Body)
		defer decoder.Decode()

		for {
			event, err := decoder.Decode()
			if err != nil {
				if err == io.EOF {
					break
				}
				continue
			}
			text := event.Data()
			if text == "" {
				continue
			}
			//g.Log().Info(ctx, "text: ", text)
			if text == "[DONE]" {
				resp.Close()
				break
			}
			// gjson.New(text).Dump()
			role := gjson.New(text).Get("message.author.role").String()
			if role == "assistant" {
				model_slug := gjson.New(text).Get("message.metadata.model_slug").String()
				if model_slug != "" {
					modelSlug = model_slug
				}
				message := gjson.New(text).Get("message.content.parts.0").String()
				msgCode2 := gjson.New(text).Get("message.content.text").String()
				if message != "" {
					content = message
					g.Log().Debug(ctx, "message: ", modelSlug, content)
				} else if msgCode2 != "" {
					msgCode = msgCode2
					g.Log().Debug(ctx, "msgCode: ", modelSlug, msgCode)
					// } else {
					// 	g.Log().Error(ctx, "可能: ", text)
				}

			} else if role == "tool" {
				g.Log().Debug(ctx, "tool: ", role, text)
				toolStr := myFileTool(ctx, text, accessToken)
				if toolStr != "" {
					msgTool = toolStr
				}
			} else {
				g.Log().Debug(ctx, "role2: ", role, text)
			}
		}
		g.Log().Debug(ctx, "最后 msgCode:", msgCode)
		g.Log().Debug(ctx, "最后 msgTool:", msgTool)
		g.Log().Debug(ctx, "最后 content:", content)
		allContent := ""
		if msgCode != "" {
			allContent += "\n```\n" + msgCode + "\n```\n"
		}
		if msgTool != "" {
			allContent += "\n" + msgTool + "\n"
		}
		content = allContent + content
		completionTokens := CountTokens(content)
		promptTokens := CountTokens(newMessages)
		totalTokens := completionTokens + promptTokens

		apiResp := gjson.New(ApiRespStr)
		apiResp.Set("choices.0.message.content", content)
		id := config.GenerateID(29)
		apiResp.Set("id", id)
		apiResp.Set("created", time.Now().Unix())
		// if req.Model == "gpt-4" {
		// 	apiResp.Set("model", "gpt-4")
		// }
		apiResp.Set("model", req.Model)

		apiResp.Set("usage.prompt_tokens", promptTokens)
		apiResp.Set("usage.completion_tokens", completionTokens)
		apiResp.Set("usage.total_tokens", totalTokens)

		if completionTokens > 0 {
			r.Response.WriteJson(apiResp)
		} else {
			g.Log().Info(ctx, "[非流]出错会话", email, realModel, "->", modelSlug)
			r.Response.Status = 501
			r.Response.WriteJson(gApiError("请重试", gerror.New("wss error")))
		}

		if realModel != "text-davinci-002-render-sha" && modelSlug == "text-davinci-002-render-sha" {
			isPlusInvalid = true

			g.Log().Info(ctx, userToken, "使用", email, icount, realModel, "->", modelSlug, "PLUS失效")
		} else {
			g.Log().Info(ctx, "[非流]完成会话", email, completionTokens, realModel, "->", modelSlug)
		}
	}

}
