// Package syobocalはしょぼいカレンダーのAPIへのアクセスを提供します
package syobocal

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	// しょぼいカレンダーのデータベースAPIのエンドポイント
	dbURL = "https://cal.syoboi.jp/db.php"
	// タイムアウト時間 (10秒)
	requestTimeout = 10 * time.Second
	userAgent      = "Annict (https://annict.com)"

	resultCodeOK = 200
	// resultCodeNotFoundは、指定したTIDの作品が1件も無いときにAPIが返すコード。
	resultCodeNotFound = 404
	// TitleLookupのTID指定は1リクエストにつき最大100件。
	titleLookupLimit = 100
)

// ClientはしょぼいカレンダーのデータベースのAPI (db.php) のクライアント
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClientは新しいしょぼいカレンダーのAPIクライアントを作成する
func NewClient() *Client {
	return &Client{
		baseURL: dbURL,
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

// titleLookupResponseはTitleLookupコマンドのレスポンス
type titleLookupResponse struct {
	Result struct {
		Code    int    `xml:"Code"`
		Message string `xml:"Message"`
	} `xml:"Result"`
	TitleItems []titleItem `xml:"TitleItems>TitleItem"`
}

type titleItem struct {
	TID       int32  `xml:"TID"`
	SubTitles string `xml:"SubTitles"`
}

// LookupSubTitlesは指定したTIDの作品のサブタイトルを、TID → 話数 → サブタイトルの形で返す。
// サブタイトルの無い作品や、しょぼいカレンダーに無いTIDは結果に含めない。
func (c *Client) LookupSubTitles(ctx context.Context, tids []int32) (map[int32]map[int32]string, error) {
	result := make(map[int32]map[int32]string)
	for batch := range slices.Chunk(tids, titleLookupLimit) {
		subTitles, err := c.lookupSubTitles(ctx, batch)
		if err != nil {
			return nil, err
		}
		for tid, titles := range subTitles {
			result[tid] = titles
		}
	}
	return result, nil
}

// lookupSubTitlesは件数上限内のTIDを1回のリクエストで取得する。
func (c *Client) lookupSubTitles(ctx context.Context, tids []int32) (map[int32]map[int32]string, error) {
	result := make(map[int32]map[int32]string)
	ids := make([]string, len(tids))
	for i, tid := range tids {
		ids[i] = strconv.FormatInt(int64(tid), 10)
	}
	query := url.Values{}
	query.Set("Command", "TitleLookup")
	query.Set("TID", strings.Join(ids, ","))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("リクエストの作成に失敗: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("しょぼいカレンダーへのリクエストに失敗: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("しょぼいカレンダーが想定外のステータスを返しました: %d", resp.StatusCode)
	}

	var body titleLookupResponse
	if err := xml.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("しょぼいカレンダーのレスポンスの解釈に失敗: %w", err)
	}

	switch body.Result.Code {
	case resultCodeOK:
	case resultCodeNotFound:
		return result, nil
	default:
		return nil, fmt.Errorf("しょぼいカレンダーがエラーを返しました: %d %s", body.Result.Code, body.Result.Message)
	}

	for _, item := range body.TitleItems {
		subTitles := parseSubTitles(item.SubTitles)
		if len(subTitles) > 0 {
			result[item.TID] = subTitles
		}
	}
	return result, nil
}

// parseSubTitlesは `*01*サブタイトル` の行の並びを、話数 → サブタイトルの形にする。
// 話数として読めない行と、サブタイトルが空の行は捨てる。
func parseSubTitles(text string) map[int32]string {
	subTitles := make(map[int32]string)
	for _, line := range strings.Split(text, "\n") {
		parts := strings.SplitN(strings.TrimRight(line, "\r"), "*", 3)
		if len(parts) != 3 || parts[0] != "" {
			continue
		}

		number, err := strconv.ParseInt(parts[1], 10, 32)
		if err != nil {
			continue
		}
		title := strings.TrimSpace(parts[2])
		if title == "" {
			continue
		}
		subTitles[int32(number)] = title
	}
	return subTitles
}
