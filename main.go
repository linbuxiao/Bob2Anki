package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"github.com/xuri/excelize/v2"
	"os"
	"strings"
	"time"
)

const (
	cliName      = "bob export to anki"
	timeLayout   = "2006-01-02 15:04:05"
	bobSheetName = "Sheet1"
)

type rowType struct {
	ActionTime time.Time
	Before     string
	After      string
}

func main() {
	app := &cli.App{
		Name: cliName,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "lastTime",
				Aliases:     []string{"lt"},
				DefaultText: "1970-01-01 00:00:00",
			},
			&cli.StringFlag{
				Name:     "filePath",
				Required: true,
				Aliases:  []string{"f"},
			},
		},
		Action: func(c *cli.Context) error {
			lastTimeStr := c.String("lastTime")
			lastTime, err := time.Parse(timeLayout, lastTimeStr)
			if err != nil {
				return err
			}
			filePath := c.String("filePath")
			rows, err := getBobExportFileByURL(filePath)
			if err != nil {
				return err
			}
			arr, err := parseRows2NameMap(rows, lastTime)
			return writeFile(arr)
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func getBobExportFileByURL(filePath string) ([][]string, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	rows, err := f.GetRows(bobSheetName)
	if err != nil {
		return nil, err
	}
	return rows[1:], nil
}

func parseRows2NameMap(rows [][]string, lastTime time.Time) ([]rowType, error) {
	var result []rowType
	for _, row := range rows {
		var r rowType
		timeStr := formatTimeStr(row[1])
		t, err := time.ParseInLocation(timeLayout, timeStr, time.Local)
		if err != nil {
			return nil, err
		}
		if t.Before(lastTime) {
			continue
		}
		r.ActionTime = t
		r.Before = row[4]
		r.After = row[6]
		result = append(result, r)
	}
	return result, nil
}

func formatTimeStr(str string) string {
	result := str
	for _, v := range []string{"年", "月", "日", "时", "分", "秒"} {
		switch v {
		case "年", "月":
			result = strings.ReplaceAll(result, v, "-")
		case "日", "秒":
			result = strings.ReplaceAll(result, v, "")
		case "时", "分":
			result = strings.ReplaceAll(result, v, ":")
		}
	}
	return result
}

func writeFile(rows []rowType) error {
	f, err := os.Create("output.txt")
	if err != nil {
		return err
	}
	for _, r := range rows {
		str := fmt.Sprintf("%s;\"%s\"\n", r.Before, r.After)
		_, err := f.WriteString(str)
		if err != nil {
			return err
		}
	}
	return nil
}
