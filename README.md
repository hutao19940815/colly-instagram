# colly-instagram
colly 采集ins博主图片、视频并下载

运行环境：golang1.14 能够科学上网

colly爬虫框架的项目入口是ins/src/ins/instagram.go 文件。 进入到ins/ins/src/ins目录，命令行执行"go run instagram.go",日志打印出"scraping finished"则代表爬取完成。下载的图片在ins/ins/src/ins/image目录下，视频在ins/ins/src/ins/video目录下。由于脚本测试的"杨紫"的instagram,博主并未视频发布，故而没有视频下载。
