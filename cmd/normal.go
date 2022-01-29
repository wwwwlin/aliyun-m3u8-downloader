package cmd

import (
	"fmt"
	"github.com/lbbniu/aliyun-m3u8-downloader/pkg/download"
	"github.com/lbbniu/aliyun-m3u8-downloader/pkg/tool"
	"github.com/spf13/cobra"
)

// normalCmd represents the normal command
var normalCmd = &cobra.Command{
	Use:   "normal",
	Short: "普通m3u8 或 标准AES-128加密 下载",
	Long: `普通m3u8 或 标准AES-128加密 下载. 使用示例:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		url, _ := cmd.Flags().GetString("url")
		output, _ := cmd.Flags().GetString("output")
		chanSize, _ := cmd.Flags().GetInt("chanSize")
		if url == "" {
			tool.PanicParameter("url")
		}
		if output == "" {
			tool.PanicParameter("output")
		}
		if chanSize <= 0 {
			panic("parameter 'chanSize' must be greater than 0")
		}

		downloader, err := download.NewTask(output, url, "")
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
	rootCmd.AddCommand(normalCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// normalCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// normalCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	normalCmd.Flags().StringP("url", "u", "", "m3u8 地址")
	normalCmd.Flags().StringP("output", "o", "", "下载保存位置")
	normalCmd.Flags().IntP("chanSize", "c", 1, "下载并发数")
	_ = normalCmd.MarkFlagRequired("url")
	_ = normalCmd.MarkFlagRequired("output")
	_ = normalCmd.MarkFlagRequired("chanSize")
}
