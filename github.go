package github

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
)

var instance *github
var logger = utils.GetModuleLogger("com.aimerneige.github")

type github struct {
}

func init() {
	instance = &github{}
	bot.RegisterModule(instance)
}

func (g *github) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       "com.aimerneige.github",
		Instance: instance,
	}
}

// Init 初始化过程
// 在此处可以进行 Module 的初始化配置
// 如配置读取
func (g *github) Init() {
}

// PostInit 第二次初始化
// 再次过程中可以进行跨 Module 的动作
// 如通用数据库等等
func (g *github) PostInit() {
}

// Serve 注册服务函数部分
func (g *github) Serve(b *bot.Bot) {
	b.GroupMessageEvent.Subscribe(func(c *client.QQClient, msg *message.GroupMessage) {
		msgString := msg.ToString()
		msgString = strings.TrimSpace(msgString)
		// 内容过短
		if len(msgString) <= 19 {
			return
		}
		// 开头不是 GitHub 域名
		if msgString[:19] != "https://github.com/" {
			return
		}
		repoInfo := msgString[19:]
		// 去除域名后内容过短
		if len(repoInfo) <= 1 {
			return
		}
		// 删除末尾的 /
		if repoInfo[len(repoInfo)-1] == '/' {
			repoInfo = repoInfo[:len(repoInfo)-1]
		}
		// 检查是否同时含有用户名和仓库名
		repoInfoSlice := strings.Split(repoInfo, "/")
		if len(repoInfoSlice) != 2 {
			return
		}
		// 获取 svg 数据
		svgURL := fmt.Sprintf("https://socialify.git.ci/%s/%s/image?description=1&font=Bitter&forks=1&issues=1&name=1&owner=1&pattern=Circuit%%20Board&pulls=1&stargazers=1&theme=Light", repoInfoSlice[0], repoInfoSlice[1])
		svgData, err := getRequest(svgURL)
		if err != nil {
			logger.WithError(err).Error("Fail to get svg data")
			return
		}
		// 如果返回 Not found 则说明仓库为私有或不存在
		if string(svgData) == "Not found" {
			c.SendGroupMessage(msg.GroupCode, message.NewSendingMessage().Append(message.NewText(msgString+"\n\n没有找到该仓库信息，该仓库可能为私有或不存在。")))
			return
		}
		// 获取图片数据
		imgURL := "https://image.thum.io/get/width/1280/crop/640/viewportWidth/1280/png/noanimate/" + svgURL
		imgData, err := getRequest(imgURL)
		if err != nil {
			logger.WithError(err).Error("Fail to get image data")
			return
		}
		// 上传图片
		uploadTarget := message.Source{
			SourceType: message.SourceGroup,
			PrimaryID:  msg.GroupCode,
		}
		uploadedImage, err := c.UploadImage(uploadTarget, bytes.NewReader(imgData))
		if err != nil {
			logger.WithError(err).Error("Fail to upload image")
			return
		}
		// 发送消息
		imgMsg := message.NewSendingMessage().Append(uploadedImage).Append(message.NewText(msgString))
		c.SendGroupMessage(msg.GroupCode, imgMsg)
	})
}

// Start 此函数会新开携程进行调用
// ```go
//
//	go exampleModule.Start()
//
// ```
// 可以利用此部分进行后台操作
// 如 http 服务器等等
func (g *github) Start(b *bot.Bot) {
}

// Stop 结束部分
// 一般调用此函数时，程序接收到 os.Interrupt 信号
// 即将退出
// 在此处应该释放相应的资源或者对状态进行保存
func (g *github) Stop(b *bot.Bot, wg *sync.WaitGroup) {
	// 别忘了解锁
	defer wg.Done()
}

func getRequest(url string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
