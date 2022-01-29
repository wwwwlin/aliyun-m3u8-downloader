package cmd

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/lbbniu/aliyun-m3u8-downloader/pkg/download"
	"github.com/lbbniu/aliyun-m3u8-downloader/pkg/request"
	"github.com/lbbniu/aliyun-m3u8-downloader/pkg/tool"
	"github.com/spf13/cobra"
	"log"
)

// aliyunCmd represents the aliyun command
var aliyunCmd = &cobra.Command{
	Use:   "aliyun",
	Short: "阿里云私有m3u8加密下载工具",
	Long: `阿里云私有m3u8加密下载工具. 使用示例:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		playAuth, _ := cmd.Flags().GetString("playAuth")
		videoId, _ := cmd.Flags().GetString("videoId")
		output, _ := cmd.Flags().GetString("output")
		chanSize, _ := cmd.Flags().GetInt("chanSize")
		if playAuth == "" {
			tool.PanicParameter("playAuth")
		}
		if videoId == "" {
			tool.PanicParameter("videoId")
		}
		if output == "" {
			tool.PanicParameter("output")
		}
		if chanSize <= 0 {
			panic("parameter 'chanSize' must be greater than 0")
		}
		// 随机字符串
		clientRand := uuid.NewString()
		sj, err := request.GetVodPlayerInfo(clientRand, playAuth, videoId)
		if err != nil {
			log.Fatalln(err)
			return
		}
		//tool.PrintJson(sj)
		playInfoList, err := sj.Get("PlayInfoList").Get("PlayInfo").Array()
		if err != nil {
			log.Fatalln(err)
			return
		}
		playInfo := sj.Get("PlayInfoList").Get("PlayInfo").GetIndex(len(playInfoList) - 1)
		tool.PrintJson(playInfo)
		serverRand, _ := playInfo.Get("Rand").String()
		plaintext, _ := playInfo.Get("Plaintext").String()
		playURL, _ := playInfo.Get("PlayURL").String()
		tool.PrintJson(playURL)
		key := tool.DecryptKey(clientRand, serverRand, plaintext)
		downloader, err := download.NewTask(output, playURL, key)
		if err != nil {
			panic(err)
		}
		if err := downloader.Start(chanSize); err != nil {
			panic(err)
		}
		fmt.Println("Done!")
	},
}

func init() {
	rootCmd.AddCommand(aliyunCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// aliyunCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// aliyunCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	aliyunCmd.Flags().StringP("playAuth", "p", "", "web播放认证信息")
	aliyunCmd.Flags().StringP("videoId", "v", "", "视频id")
	aliyunCmd.Flags().StringP("output", "o", "", "下载保存位置")
	aliyunCmd.Flags().IntP("chanSize", "c", 1, "下载并发数")
	_ = aliyunCmd.MarkFlagRequired("playAuth")
	_ = aliyunCmd.MarkFlagRequired("videoId")
	_ = aliyunCmd.MarkFlagRequired("output")
	_ = aliyunCmd.MarkFlagRequired("chanSize")
}
