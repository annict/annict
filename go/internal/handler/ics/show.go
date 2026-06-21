package ics

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/ical"
	"github.com/annict/annict/go/internal/usecase"
)

// Show はiCalendar形式のカレンダーを返します
// GET /@{username}/ics - メインのエンドポイント
// GET /ics?username={username} - Apple カレンダー互換の代替パス
func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// URLパラメータまたはクエリパラメータからusernameを取得
	username := chi.URLParam(r, "username")
	if username == "" {
		username = r.URL.Query().Get("username")
	}

	if username == "" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// カレンダーデータを取得
	result, err := h.getUserCalendarUC.Execute(ctx, usecase.GetUserCalendarInput{
		Username: username,
		Now:      time.Now(),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		slog.ErrorContext(ctx, "カレンダーデータの取得に失敗しました", "error", err, "username", username)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	userCalendar := result.UserCalendar

	// ユーザーのロケールに基づいて作品タイトルを選択する関数
	selectTitle := func(title, titleEn string) string {
		if userCalendar.Locale == "en" && titleEn != "" {
			return titleEn
		}
		return title
	}

	// iCalendarオブジェクトを構築
	cal := &ical.Calendar{
		TimeZone: userCalendar.TimeZone,
		CalName:  fmt.Sprintf("Annict@%s", userCalendar.Username),
		Events:   make([]ical.Event, 0, len(userCalendar.Slots)+len(userCalendar.Works)),
	}

	// 放送枠をイベントに変換
	for _, slot := range userCalendar.Slots {
		workTitle := selectTitle(slot.WorkTitle, slot.WorkTitleEn)

		// サマリーを構築: "作品タイトル 話数 サブタイトル (チャンネル名)"
		summary := workTitle
		if slot.EpisodeNumber != "" {
			summary = fmt.Sprintf("%s %s", summary, slot.EpisodeNumber)
		}
		if slot.EpisodeTitle != "" {
			summary = fmt.Sprintf("%s %s", summary, slot.EpisodeTitle)
		}
		if slot.ChannelName != "" {
			summary = fmt.Sprintf("%s (%s)", summary, slot.ChannelName)
		}

		// 説明を構築: "作品タイトル 話数 サブタイトル\nエピソードURL"
		description := workTitle
		if slot.EpisodeNumber != "" {
			description = fmt.Sprintf("%s %s", description, slot.EpisodeNumber)
		}
		if slot.EpisodeTitle != "" {
			description = fmt.Sprintf("%s %s", description, slot.EpisodeTitle)
		}
		description = fmt.Sprintf("%s\nhttps://%s/works/%d/episodes/%d", description, h.cfg.Domain, slot.WorkID, slot.EpisodeID)

		cal.Events = append(cal.Events, ical.Event{
			UID:         fmt.Sprintf("slot-%d@annict.com", slot.ID),
			Summary:     summary,
			Description: description,
			Start:       slot.StartedAt,
			End:         slot.StartedAt.Add(30 * time.Minute), // 放送枠は30分と仮定
			AllDay:      false,
		})
	}

	// 作品（放送開始日）をイベントに変換
	for _, work := range userCalendar.Works {
		workTitle := selectTitle(work.Title, work.TitleEn)
		description := fmt.Sprintf("%s\nhttps://%s/works/%d", workTitle, h.cfg.Domain, work.ID)

		cal.Events = append(cal.Events, ical.Event{
			UID:         fmt.Sprintf("work-%d@annict.com", work.ID),
			Summary:     workTitle,
			Description: description,
			Start:       work.StartedOn,
			End:         work.StartedOn.AddDate(0, 0, 1), // 終日イベントは翌日を終了日に設定
			AllDay:      true,
		})
	}

	// ICS形式でレスポンス
	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="annict.ics"`)
	w.WriteHeader(http.StatusOK)

	// A failed write here means the response body could not be delivered to the
	// client: the connection was closed or the write deadline was exceeded (e.g.
	// a calendar importer that opened the request then stopped reading). Once the
	// header is sent, a w.Write error always originates from the underlying
	// net.Conn, so it is a client-side/transport failure, not a server error.
	// Log it at warn level: the slog Sentry handler captures only Error and
	// Fatal, so warn keeps this transport noise out of Sentry while still
	// recording it in the local structured logs.
	//
	// [Ja] ここでの書き込み失敗は、レスポンスボディをクライアントに送り切れな
	// かったことを意味する。接続が閉じられたか書き込みデッドラインを超過した
	// ケース (例: リクエストを開いたまま読み取りをやめたカレンダーインポーター)
	// である。ヘッダー送出後の w.Write エラーは常に背後の net.Conn に起因する
	// ため、サーバーエラーではなくクライアント側・トランスポート由来の失敗で
	// ある。このため warn レベルでログ出力する。slog の Sentry ハンドラーは
	// Error と Fatal のみをイベント化するため、warn にすればローカルの構造化
	// ログには残しつつ、このトランスポートノイズを Sentry に送らずに済む。
	if _, err := w.Write([]byte(cal.ToICS())); err != nil {
		slog.WarnContext(ctx, "ICSデータの書き込みに失敗しました", "error", err)
	}
}
