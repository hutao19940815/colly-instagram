package main

import (
	"fmt"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"strings"
	"encoding/json"
	"os"
	"bufio"
	"io"
	"net/http"
	"strconv"
)

type pageInfo struct {
	EndCursor string `json:"end_cursor"`
	NextPage  bool   `json:"has_next_page"`
}

type tomedia struct {
	PageInfo pageInfo `json:"page_info"`
	Edges[] struct{
		Node struct{
			//Id  string `json:"id"`
			Shortcode string `json:"shortcode"`
			ImageUrl string `json:"display_url"`
			IsVideo bool `json:"is_video"`
			Date int `json:"taken_at_timestamp"`
			TypeName string `json:"__typename"`
		}`json:"node"`
	}`json:"edges"`
}

type Resp struct {
	EntryData struct{
		ProfilePage[] struct{
			Graphql struct{
				User struct {
					UserId string `json:"id"`
					Edge_owner_to_timeline_media tomedia `json:"edge_owner_to_timeline_media"`
				}`json:"user"`
			} `json:"graphql"`
		}`json:"ProfilePage"`
	}`json:"entry_data"`
}


type Xhr struct {
	Data  struct{
		User struct{
			Edge_timeline tomedia `json:"edge_owner_to_timeline_media"`
		}`json:"user"`
	}`json:"data"`
}

type Video struct {
	Graphq  struct{
		Shortcode_media struct{
			Video_url string `json:"video_url"`
		}`json:"shortcode_media"`
	}`json:"graphq"`
}

//下载图片和视频
func download(url string,save_path string,num int)int{

	res, err :=http.Get(url)
	if err != nil {
		fmt.Println("A error occurred!")
		return num
	}
	reader := bufio.NewReaderSize(res.Body, 32 * 1024)
	file, err := os.Create(save_path)
	if err != nil {
		panic(err)
	}
	writer := bufio.NewWriter(file)
	//
	io.Copy(writer, reader)
	num+=1
	return num
}
//下拉数据异步请求接口
const uri string =`https://www.instagram.com/graphql/query/?query_hash=e769aa130647d2354c40ea6a439bfc08&variables={"id":"%s","first":12,"after":"%s"}`
func main() {

	var user_id string   //ins博主id
	img_num := 1        //图片编号
	imgDIr := fmt.Sprintf("./image/")   //图片存储路径
	video_num := 1      //视频编号
	videoDir := fmt.Sprintf("./video/")  //视频存储路径
	c := colly.NewCollector(func(c *colly.Collector) {
		extensions.RandomUserAgent(c) // 设置随机头
		c.Async=true
	},
	) // 创建收集器

	c.OnHTML("body", func(e *colly.HTMLElement) {

		//判断页面是否是视频详情页，是则是详情页，否则为主页
		if strings.Contains(e.Text,"video_url"){

			str_1:=strings.Split(e.Text, "graphql")[1]
			str_2:=strings.Split(str_1,"]},\"hostname")[0]
			str_3:="{\"graphq"+str_2

			var  video_json Video     //截取获得详情页相关json
			err := json.Unmarshal([]byte(str_3), &video_json)
			if err != nil {
				fmt.Println(err)
			}

			video_url := video_json.Graphq.Shortcode_media.Video_url   //解析json得到视频地址
			save_path := videoDir + strconv.Itoa(video_num) + ".mp4"   //视频文件路径设置
			video_num=download(video_url,save_path,video_num)
			//fmt.Println(video_url)
		}else {

			jsonData := strings.Split(strings.Split(e.Text, "sharedData =")[1], ";")[0]

			var resp Resp  //截取获得主页相关json
			err := json.Unmarshal([]byte(jsonData), &resp)
			if err != nil {
				fmt.Println(err)
			}

			var user = resp.EntryData.ProfilePage[0].Graphql.User
			page_info := user.Edge_owner_to_timeline_media.PageInfo
			end_cursor := page_info.EndCursor   //uri中after变量值
			has_next_page := page_info.NextPage   //判断是否还有下一页
			//fmt.Println(has_next_page)
			user_id = user.UserId
			edges := user.Edge_owner_to_timeline_media.Edges
			for edge := range edges {
				node := edges[edge].Node
				shortcode := node.Shortcode
				html_url := "https://www.instagram.com/p/" + shortcode   //内容详情页地址
				if node.IsVideo {
					c.Visit(html_url)
				}
				img_url := node.ImageUrl   //解析json得到图片地址
				//fmt.Println(img_url)
				save_path := imgDIr + strconv.Itoa(img_num) + ".jpg"    //图片文件路径设置
				img_num = download(img_url, save_path, img_num)
				c.Visit(img_url)
			}
			if has_next_page {
				next_xhr_url := fmt.Sprintf(uri, user_id, end_cursor)  //完整下拉数据异步请求接口
				//fmt.Println(next_xhr_url)
				c.Visit(next_xhr_url)
			}else {
				fmt.Println("scraping finished")
			}
		}

	})

	c.OnResponse(func(r *colly.Response){
		//判断是否为请求图片响应
		if strings.Index(r.Headers.Get("Content-Type"), "image") > -1 {
			return
		}
		//判断是否为json响应
		if strings.Index(r.Headers.Get("Content-Type"), "json") == -1 {
			return
		}

		data := Xhr{}    //下拉的json数据

		err :=json.Unmarshal(r.Body, &data)
		if err != nil {
			fmt.Println(err)
		}
		//fmt.Println(data)
		var user_xhr=data.Data.User


		page_info_xhr:=user_xhr.Edge_timeline.PageInfo
		end_cursor_xhr:=page_info_xhr.EndCursor
		has_next_page_xhr:=page_info_xhr.NextPage
		edges_xhr:=user_xhr.Edge_timeline.Edges
		for edge_xhr := range  edges_xhr{
			node_xhr:=edges_xhr[edge_xhr].Node
			shortcode := node_xhr.Shortcode
			html_url := "https://www.instagram.com/p/" + shortcode   //内容详情页地址
			if node_xhr.IsVideo {
				c.Visit(html_url)
			}
			img_url:=node_xhr.ImageUrl
			//fmt.Println(img_url)
			save_path:=imgDIr+strconv.Itoa(img_num)+".jpg"
			img_num=download(img_url,save_path,img_num)
			c.Visit(img_url)
		}

		if has_next_page_xhr {
			next_xhr_url:=fmt.Sprintf(uri, user_id, end_cursor_xhr)
			//fmt.Println(next_xhr_url)
			c.Visit(next_xhr_url)
		}else {
			fmt.Println("scraping finished")
		}
	})


	c.OnError(func(response *colly.Response, err error) {
		fmt.Println(err)
	})

	c.Visit("https://www.instagram.com/yangzi_official/")   //杨紫
	c.Wait()
}