package chat

import (
	"chatgpt-api-server/apireq"
	"chatgpt-api-server/apirespstream"
	"chatgpt-api-server/config"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gogf/gf/container/garray"
	"github.com/gogf/gf/text/gregex"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/encoding/gurl"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	jsoniter "github.com/json-iterator/go"
)

func myFileTool(ctx context.Context, text string, accessToken string) string {
	//arr := gjson.New(text).Get("message.content.parts").Array() //message.content.parts
	rz := ""
	down := ""
	patternStr := `file-service://([\w\-]+)`
	//patternStr := `file-service://file-(\w+)`
	result, err := gregex.MatchAllString(patternStr, text)
	if err != nil {
		g.Log().Warning(ctx, "g.getFile().Get error: ", err)
		//return '';
	} else {
		//g.Log().Error(ctx, "有货") //s := garray.NewStrArray()
		//g.Dump(result)
		arr2 := garray.NewStrArray()
		for _, v := range result {
			arr2.Append(v[0])
		}
		arr2.Unique()
		//g.Dump(arr2)
		for i2 := 0; i2 < arr2.Len(); i2++ {
			asset_pointer := arr2.At(i2) //gjson.New(arr[i2]).Get("asset_pointer").String()
			if asset_pointer != "" {
				download_url := getFile(ctx, accessToken, asset_pointer)
				g.Log().Info(ctx, "图片: ", asset_pointer, download_url)
				if download_url != "" {
					//`![img${i+1}](${ await downloadImg(v.asset_pointer) }) `
					rz += fmt.Sprintf(" ![img%d](%s) ", i2+1, download_url) //`![img${i+1}](`+ download_url+")\n\n"
					down += fmt.Sprintf(" [下载%d](%s) ", i2+1, download_url)
				}
			}
		}
	}
	// if len(arr) > 0 {
	// 	for i2 := 0; i2 < len(arr); i2++ {
	// 		asset_pointer := gjson.New(arr[i2]).Get("asset_pointer").String()
	// 		if asset_pointer != "" {
	// 			download_url := getFile(ctx, accessToken, asset_pointer)
	// 			g.Log().Error(ctx, "图片2: ", asset_pointer, download_url)
	// 			if download_url != "" {
	// 				//`![img${i+1}](${ await downloadImg(v.asset_pointer) }) `
	// 				rz += fmt.Sprintf(" ![img%d](%s) ", i2+1, download_url) //`![img${i+1}](`+ download_url+")\n\n"
	// 				down += fmt.Sprintf(" [下载%d](%s) ", i2+1, download_url)
	// 			}
	// 		}
	// 	}
	// }
	return rz + down
}

//backend-api/files/${file.replace('file-service://','')}/download
//https://free.xyhelper.com.cn/backend-api/files/file-cxzEOSseNPAtmLHTWBPFT83G/download

func getFile(ctx context.Context, accessToken string, file string) string {
	resp, err := g.Client().SetHeaderMap(g.MapStrStr{
		"Authorization": "Bearer " + accessToken,
		"Content-Type":  "application/json",
		"authkey":       config.AUTHKEY(ctx),
	}).Get(ctx, config.CHATPROXY(ctx)+"/backend-api/files/"+gstr.Replace(file, "file-service://", "")+"/download")
	if err != nil {
		g.Log().Warning(ctx, "g.getFile().Get error: ", err)
		// r.Response.Status = 500
		// r.Response.WriteJson(gjson.New(`{"detail": "internal server error"}`))
		return ""
	}
	defer resp.Close()
	if resp.StatusCode != 200 {
		g.Log().Warning(ctx, "g.getFile().Get resp.StatusCode: ", resp.StatusCode)
		return ""
	}
	download_url := gjson.New(resp.ReadAllString()).Get("download_url").String()
	if download_url == "" {
		g.Log().Warning(ctx, "get download_url fail")
		return ""
	}
	download_url = "https://wsrv.nl/?url=" + gurl.Encode(download_url)
	return download_url

}

func toStream(ctx context.Context, id string, content string, req *apireq.Req, r *ghttp.Request) {
	apiResp := gjson.New(ApiRespStrStream)
	apiResp.Set("id", id)
	apiResp.Set("created", time.Now().Unix())
	apiResp.Set("choices.0.delta.content", content)
	apiResp.Set("model", req.Model)
	apiRespStruct := &apirespstream.ApiRespStreamStruct{}
	gconv.Struct(apiResp, apiRespStruct)
	// g.Dump(apiRespStruct)
	// 创建一个jsoniter的Encoder
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	// 将结构体转换为JSON文本并保持顺序
	sortJson, err := json.Marshal(apiRespStruct)
	if err != nil {
		g.Log().Error(ctx, "转换JSON出错:", err)
		return
	}
	r.Response.Writeln("data: " + string(sortJson) + "\n\n")
	r.Response.Flush()

}
func Myfun(r *ghttp.Request) {
	r.Response.Header().Set("Access-Control-Allow-Origin", "*")
	r.Response.Header().Set("Access-Control-Allow-Headers", "*")
	r.Response.Header().Set("Access-Control-Allow-Credentials", "true")
	r.Response.Header().Set("Access-Control-Allow-Methods", "*")
	r.Response.Header().Set("Access-Control-Max-Age", "3600")
	r.Response.Status = 200
	r.Response.WriteExit("ok")

}

func gApiError(message string, e error) (err *gjson.Json) {
	err = gjson.New(apiError)
	err.Set("error.message", message)
	err.Set("error.param", e.Error())
	return
}

func gpt4miz(ctx g.Ctx, messages string, accessToken string, ChatReq *gjson.Json) (ChatReqOut *gjson.Json, err error) {
	patternStr := `\b(?:https?):\/\/[-A-Za-z0-9+&@#\/%?=~_|!:,.;]+[-A-Za-z0-9+&@#\/%=~_|]`
	result, err := gregex.MatchAllString(patternStr, messages)
	if err != nil {
		return
	}
	filename := ""
	fileType := ""
	file_id := ""
	size_bytes := int64(0)
	file_size_tokens := int64(-1)
	newMessages := ""
	ChatReqOut = ChatReq
	for i, v := range result {
		//只搞3个文件
		if i > 2 {
			break
		}
		//arr2.Append(v[0])

		filename, fileType, err = downloadFile(ctx, v[0])
		if err != nil {
			return
		}

		file_id, size_bytes, file_size_tokens, err = myUploadAzure(ctx, "./temp/"+filename, accessToken, "gpt-4-gizmo")
		if err != nil {
			return
		}
		ChatReq.Set("messages.0.metadata.attachments."+gconv.String(i)+".name", filename)
		ChatReq.Set("messages.0.metadata.attachments."+gconv.String(i)+".id", file_id)
		ChatReq.Set("messages.0.metadata.attachments."+gconv.String(i)+".mimeType", fileType)
		ChatReq.Set("messages.0.metadata.attachments."+gconv.String(i)+".size", size_bytes)
		if file_size_tokens > 0 {
			ChatReq.Set("messages.0.metadata.attachments."+gconv.String(i)+".fileTokenSize", file_size_tokens)
		}

	}

	newMessages, err = gregex.ReplaceString(patternStr, "", messages)
	if err != nil {
		return
	}
	ChatReq.Set("messages.0.content.parts.0", newMessages)
	ChatReqOut = ChatReq

	return
}

func gpt4all(ctx g.Ctx, messages string, accessToken string) (ChatReq *gjson.Json, newMessages string, err error) {
	ChatReq = gjson.New(chat4allReq)
	patternStr := `\b(?:https?):\/\/[-A-Za-z0-9+&@#\/%?=~_|!:,.;]+[-A-Za-z0-9+&@#\/%=~_|]`
	result, err := gregex.MatchAllString(patternStr, messages)
	if err != nil {
		return
	}
	oneUrl := gstr.ToLower(result[0][0])
	if gstr.Contains(oneUrl, ".jpg") || gstr.Contains(oneUrl, ".png") || gstr.Contains(oneUrl, ".gif") || gstr.Contains(oneUrl, ".jpeg") || gstr.Contains(oneUrl, ".webp") || gstr.Contains(oneUrl, ".avif") {
		g.Log().Info(ctx, "走图片模式！")
		newMessages, err = gregex.ReplaceString(patternStr, "", messages)
		if err != nil {
			return
		}
		ChatReq, err = gpt4vReq(ctx, oneUrl, accessToken, newMessages)
		return
	}
	filename := ""
	fileType := ""
	file_id := ""
	size_bytes := int64(0)
	file_size_tokens := int64(-1)
	//var filenames []string

	for i, v := range result {
		//只搞3个文件
		if i > 2 {
			break
		}
		//arr2.Append(v[0])

		filename, fileType, err = downloadFile(ctx, v[0])
		if err != nil {
			return
		}

		file_id, size_bytes, file_size_tokens, err = myUploadAzure(ctx, "./temp/"+filename, accessToken, "gpt-4-all")
		if err != nil {
			return
		}
		ChatReq.Set("messages.0.metadata.attachments."+gconv.String(i)+".name", filename)
		ChatReq.Set("messages.0.metadata.attachments."+gconv.String(i)+".id", file_id)
		ChatReq.Set("messages.0.metadata.attachments."+gconv.String(i)+".mimeType", fileType)
		ChatReq.Set("messages.0.metadata.attachments."+gconv.String(i)+".size", size_bytes)
		if file_size_tokens > 0 {
			ChatReq.Set("messages.0.metadata.attachments."+gconv.String(i)+".fileTokenSize", file_size_tokens)
		}

	}

	newMessages, err = gregex.ReplaceString(patternStr, "", messages)
	if err != nil {
		return
	}
	ChatReq.Set("messages.0.content.parts.0", newMessages)
	// g.Dump(ChatReq)
	// g.Dump(result)
	// g.Dump(newMessages)
	return
}
func downloadFile(ctx g.Ctx, urlString string) (outputPath string, fileType string, err error) {
	// 检查 ./temp 目录是否存在 不在则创建

	// , outputPath string
	//outputPath= url.Parse(urlString);
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		//g.Log().Error("URL 解析失败:", err)
		return
	}

	outputPath = path.Base(parsedURL.Path)
	response, err := http.Get(urlString)
	if err != nil {
		return
	}
	defer response.Body.Close()

	if outputPath == "" || outputPath == "/" {
		outputPath = config.GenerateID(7) + ".html"
	}

	file, err := os.Create("./temp/" + outputPath)
	if err != nil {
		return
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return
	}
	mime, err := mimetype.DetectFile("./temp/" + outputPath)
	fileType = mime.String()
	//g.Log().Info(ctx, "下载文件", urlString, fi, fileType)
	return
}

func myUploadAzure(ctx g.Ctx, filepath string, token string, model string) (file_id string, size_bytes int64, file_size_tokens int64, err error) {
	// 检测文件是否存在
	if !gfile.Exists(filepath) {
		err = gerror.New("read file fail")
		return
	}
	// 删除临时文件
	defer gfile.Remove(filepath)

	fileName := gfile.Basename(filepath)
	fileSize := gfile.Size(filepath)

	// 获取上传地址 backend-api/files  POST
	//res, err := g.Client().SetHeader("Authorization", "Bearer "+token).ContentJson().Post(ctx, config.CHATPROXY(ctx)+"/backend-api/files", g.Map{
	//Chatgpt-Account-Id: 3cdd0789-ba9e-434d-b60b-dabe75dd8f87
	//3cdd0789-ba9e-434d-b60b-dabe75dd8f87
	res, err := g.Client().SetHeaderMap(g.MapStrStr{
		"Authorization": "Bearer " + token,
		//"Chatgpt-Account-Id": "3cdd0789-ba9e-434d-b60b-dabe75dd8f87",
	}).ContentJson().Post(ctx, config.CHATPROXY(ctx)+"/backend-api/files", g.Map{
		"file_name": fileName,
		"file_size": fileSize,
		//"use_case":  "multimodal",
		"use_case": "my_files",
	})
	if err != nil {
		return
	}
	defer res.Close()
	if res.StatusCode != 200 {
		res.RawDump()
		// statusCode = res.StatusCode
		err = gerror.New("get upload_url fail:" + res.Status)
		return
	}

	//
	resJson := gjson.New(res.ReadAllString())
	//resJson.Dump()
	upload_url := resJson.Get("upload_url").String()
	file_id = resJson.Get("file_id").String()
	if upload_url == "" {
		err = gerror.New("get upload_url fail")
		return
	}
	//g.Log().Info(ctx, "上传图片", fileName, file_id, upload_url)

	size_bytes = fileSize

	// 以二进制流的方式上传文件 PUT
	filedata := gfile.GetBytes(filepath)

	resput, err := g.Client().SetHeaderMap(g.MapStrStr{
		"x-ms-blob-type": "BlockBlob",
		"x-ms-version":   "2020-04-08",
	}).Put(ctx, upload_url, filedata)
	if err != nil {
		return
	}
	defer resput.Close()
	// resput.RawDump()
	if resput.StatusCode != 201 {
		err = gerror.New("upload file fail")
		return
	}
	// 获取文件下载地址 backend-api/files/{file_id}/uploaded  POST
	resdown, err := g.Client().SetHeader("Authorization", "Bearer "+token).ContentJson().Post(ctx, config.CHATPROXY(ctx)+"/backend-api/files/"+file_id+"/uploaded")
	if err != nil {
		return
	}
	defer resdown.Close()
	//resdown.RawDump()
	download_url := gjson.New(resdown.ReadAllString()).Get("download_url").String()
	if download_url == "" {
		err = gerror.New("get download_url fail")
		return
	}
	g.Log().Info(ctx, "上传", fileName, file_id, download_url)
	if model != "gpt-4-all" { //只有gpt-4-all 这个模型才需要确认 其他不需要确认
		file_size_tokens = -10
		return
	}
	file_size_tokens = -1
	i := 1
	for i <= 15 {
		file_size_tokens, err = uploadFileCheck(ctx, token, file_id)
		if err != nil {
			return
		}
		if file_size_tokens > -1 || file_size_tokens == -11 { //忽略的
			break
		}

		time.Sleep(200 * time.Microsecond)
		i++
	}
	if file_size_tokens == -1 {
		g.Log().Info(ctx, "上传解析失败 ")
		err = gerror.New("文件解析失败")
		return
	}
	return
}

func uploadFileCheck(ctx g.Ctx, token string, file_id string) (file_size_tokens int64, err error) {
	// {
	// 	"id": "file-rHWA5hR9nghYGZnXbjPpcQS3",
	// 	"name": "asdfads.txt",
	// 	"creation_time": "2024-01-30 14:11:34.865116+00:00",
	// 	"state": "ready",
	// 	"ready_time": "2024-01-30 14:11:40.740552+00:00",
	// 	"size": 19,
	// 	"metadata": {
	// 		"retrieval": {
	// 			"status": "success",
	// 			"file_size_tokens": 5
	// 		}
	// 	},
	// 	"use_case": "my_files",
	// 	"retrieval_index_status": "success",
	// 	"file_size_tokens": 5,
	// 	"variants": null
	// }
	resV, err := g.Client().SetHeader("Authorization", "Bearer "+token).ContentJson().Get(ctx, config.CHATPROXY(ctx)+"/backend-api/files/"+file_id)
	file_size_tokens = -1
	if err != nil {
		return
	}
	defer resV.Close()
	resJson := gjson.New(resV.ReadAllString())
	//resJson.Dump() // resJson.Get("state").String() == "ready" &&
	statsu := resJson.Get("retrieval_index_status").String()
	if statsu == "success" {
		file_size_tokens = resJson.Get("file_size_tokens").Int64()
		g.Log().Info(ctx, "上传成功解析！", file_size_tokens)
	} else {
		g.Log().Info(ctx, "解析", statsu)
	}
	if statsu == "skipped" {
		file_size_tokens = -11
		return
	}

	if statsu == "failed" {
		file_size_tokens = resJson.Get("file_size_tokens").Int64()
		g.Log().Info(ctx, "文件解析失败", file_size_tokens)
		err = gerror.New("文件解析失败")
		return
	}

	return
}

func isChatGPtUrl(newMessages string) (isChat bool) {
	count := gstr.Count(newMessages, "http")
	isChat = false
	if count == 0 {
		return
	}
	count = gstr.Count(newMessages, "oaiusercontent")
	if count > 0 {
		isChat = true
		return
	}
	return
}
