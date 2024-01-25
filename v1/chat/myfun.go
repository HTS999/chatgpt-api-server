package chat

import (
	"chatgpt-api-server/apireq"
	"chatgpt-api-server/apirespstream"
	"chatgpt-api-server/config"
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/container/garray"
	"github.com/gogf/gf/text/gregex"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/encoding/gurl"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
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
