package db_person

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authMiddleware "github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/testutil"
)

// TestIndexはDB人物一覧ページを描画し、公開中とアーカイブの人物を誰にでも表示し、閲覧者の
// ロールに応じて操作を出し分けることを、リクエストコンテキストのユーザーからend-to-endで検証する。
func TestIndex(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	publishedID := testutil.NewPersonBuilder(t, tx).
		WithName("公開中の人物-" + t.Name()).
		WithNameKana("こうかいちゅうのじんぶつ").
		WithNameEn("Published Person").
		Build()
	archivedID := testutil.NewPersonBuilder(t, tx).
		WithName("アーカイブの人物-" + t.Name()).
		WithStatus(model.PersonStatusArchived).
		Build()
	testutil.NewPersonBuilder(t, tx).
		WithName("掲載停止の人物-" + t.Name()).
		WithStatus(model.PersonStatusDelisted).
		Build()
	testutil.NewPersonBuilder(t, tx).
		WithName("削除済みの人物-" + t.Name()).
		WithStatus(model.PersonStatusDeleted).
		Build()

	handler := newTestHandler(t, db, tx)

	publishedName := "公開中の人物-" + t.Name()
	archivedName := "アーカイブの人物-" + t.Name()
	delistedName := "掲載停止の人物-" + t.Name()
	deletedName := "削除済みの人物-" + t.Name()
	newLink := `href="/db/people/new"`
	editLink := fmt.Sprintf(`href="/db/people/%d/edit"`, int64(publishedID))
	archivedEditLink := fmt.Sprintf(`href="/db/people/%d/edit"`, int64(archivedID))
	archiveLink := fmt.Sprintf(`href="/db/people/%d/archive/new"`, int64(publishedID))
	unarchiveButton := fmt.Sprintf(`hx-delete="/db/people/%d/archive"`, int64(archivedID))
	deleteButton := fmt.Sprintf(`hx-delete="/db/people/%d"`, int64(publishedID))

	tests := []struct {
		name        string
		user        *model.User
		wantPresent []string
		wantAbsent  []string
	}{
		{
			name:        "管理者: 削除を含む操作を表示",
			user:        &model.User{ID: 1, Role: model.RoleAdmin},
			wantPresent: []string{publishedName, archivedName, newLink, editLink, archivedEditLink, archiveLink, unarchiveButton, deleteButton},
			wantAbsent:  []string{delistedName, deletedName},
		},
		{
			name:        "編集者: 削除を除く操作を表示",
			user:        &model.User{ID: 1, Role: model.RoleEditor},
			wantPresent: []string{publishedName, archivedName, newLink, editLink, archivedEditLink, archiveLink, unarchiveButton},
			wantAbsent:  []string{delistedName, deletedName, deleteButton},
		},
		{
			name:        "一般ユーザー: 操作なし",
			user:        &model.User{ID: 1, Role: model.RoleUser},
			wantPresent: []string{publishedName, archivedName},
			wantAbsent:  []string{delistedName, deletedName, newLink, editLink, archiveLink, unarchiveButton, deleteButton},
		},
		{
			name:        "未ログイン: 操作なし",
			user:        nil,
			wantPresent: []string{publishedName, archivedName},
			wantAbsent:  []string{delistedName, deletedName, newLink, editLink, archiveLink, unarchiveButton, deleteButton},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/db/people", nil)
			if tt.user != nil {
				req = req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, tt.user))
			}
			rr := httptest.NewRecorder()

			handler.Index(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
			}
			if ct := rr.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
				t.Errorf("Content-Type = %q、期待値 = text/html; charset=utf-8", ct)
			}

			body := rr.Body.String()
			for _, want := range tt.wantPresent {
				if !strings.Contains(body, want) {
					t.Errorf("期待する文字列が含まれていません: %q", want)
				}
			}
			for _, notWant := range tt.wantAbsent {
				if strings.Contains(body, notWant) {
					t.Errorf("含まれてはいけない文字列が含まれています: %q", notWant)
				}
			}
		})
	}

	t.Run("ページの共通要素", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/db/people", nil)
		rr := httptest.NewRecorder()

		handler.Index(rr, req)

		body := rr.Body.String()
		for _, expected := range []string{
			"<title>人物 | Annict DB</title>",
			fmt.Sprintf(`href="/people/%d"`, int64(publishedID)),
			"こうかいちゅうのじんぶつ",
			"Published Person",
			`<meta property="og:url" content="https://test.annict.com/db/people">`,
		} {
			if !strings.Contains(body, expected) {
				t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
			}
		}
		// robots.txtが /db/ をDisallowしているため、管理画面はcanonicalを宣言しない。
		if strings.Contains(body, `rel="canonical"`) {
			t.Error("/dbの画面にcanonicalが含まれてはいけません")
		}
	})
}

// TestIndex_OGURLは、og:urlがページ番号だけを保ち、1ページ目と不正なページ番号では
// パラメータを付けず、未知のパラメータを載せないことを検証する。
func TestIndex_OGURL(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	tests := []struct {
		name   string
		target string
		want   string
	}{
		{name: "1ページ目", target: "/db/people?page=1", want: "https://test.annict.com/db/people"},
		{name: "2ページ目", target: "/db/people?page=2&utm_source=x", want: "https://test.annict.com/db/people?page=2"},
		{name: "不正なページ番号", target: "/db/people?page=abc", want: "https://test.annict.com/db/people"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.target, nil)
			rr := httptest.NewRecorder()

			handler.Index(rr, req)

			want := fmt.Sprintf(`<meta property="og:url" content="%s">`, tt.want)
			if !strings.Contains(rr.Body.String(), want) {
				t.Errorf("og:urlが期待値と一致しません: %q", want)
			}
		})
	}
}
