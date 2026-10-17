package legal_document

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/legaldoc"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/templates/layouts"
	legalDocumentTemplate "github.com/annict/annict/go/internal/templates/pages/legal_document"
	"github.com/annict/annict/go/internal/viewmodel"
)

// documentPageは文書の種類ごとに異なるページの設定
type documentPage struct {
	titleKey   string
	headingKey string
	// noIndex・noFollow・noArchiveは <meta name="robots"> の指定。noIndexのときはcanonicalも出力しない
	noIndex   bool
	noFollow  bool
	noArchive bool
}

// documentPagesは文書の種類ごとのページの設定。
// 文書の一覧表 (legaldoc) に新しい種類の文書を登録するときは、ここにも設定を足す
var documentPages = map[legaldoc.Key]documentPage{
	legaldoc.KeyTerms: {
		titleKey:   "legal_document_show_terms_title",
		headingKey: "legal_document_show_terms_heading",
	},
	legaldoc.KeyAPITerms: {
		titleKey:   "legal_document_show_api_terms_title",
		headingKey: "legal_document_show_api_terms_heading",
	},
	legaldoc.KeyPrivacyPolicy: {
		titleKey:   "legal_document_show_privacy_policy_title",
		headingKey: "legal_document_show_privacy_policy_heading",
	},
	// 特定商取引法に基づく表記は運営者個人の氏名を載せる法定の表示で、検索結果に出す必要が無いため、
	// インデックス・リンクの追跡・キャッシュをさせない
	legaldoc.KeySpecifiedCommercialTransactions: {
		titleKey:   "legal_document_show_specified_commercial_transactions_title",
		headingKey: "legal_document_show_specified_commercial_transactions_heading",
		noIndex:    true,
		noFollow:   true,
		noArchive:  true,
	},
}

// Showは法的文書のページを表示します (GET /terms・/privacy・/legalなど)。
// 改定前の版のアーカイブページ (GET /terms/archive/20140310など) もここで表示する。
// 配信するパスは文書の一覧表 (legaldoc) で決まり、一覧表に無いパスは404を返す。
func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	if doc, ok := legaldoc.Lookup(r.URL.Path); ok {
		h.showDocument(w, r, doc)
		return
	}

	if version, doc, ok := legaldoc.LookupVersion(r.URL.Path); ok {
		h.showVersion(w, r, version, doc)
		return
	}

	httperror.NotFound(w, r)
}

// showDocumentは文書の現行の版を表示する
func (h *Handler) showDocument(w http.ResponseWriter, r *http.Request, doc legaldoc.Document) {
	// 文書のページの言語はURLで決まる。リクエストのロケールを文書の言語で上書きし、
	// サイドバー・フッター・タイトルを含むページ全体を本文と同じ言語で描画する
	ctx := i18n.SetLocale(r.Context(), doc.Lang)

	page, ok := documentPages[doc.Key]
	if !ok {
		slog.ErrorContext(ctx, "法的文書のページの設定が登録されていない", "document_key", doc.Key)
		httperror.InternalServerError(w, r.WithContext(ctx))
		return
	}

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg, doc.Path)
	meta.SetTitle(ctx, page.titleKey)
	meta.Locale = doc.Lang
	meta.NoIndex = page.noIndex
	meta.NoFollow = page.noFollow
	meta.NoArchive = page.noArchive
	versions := legaldoc.LanguageVersions(doc.Key)
	languageVersions := make([]legalDocumentTemplate.LanguageVersion, 0, len(versions))
	for _, version := range versions {
		meta.AddLanguageVersion(h.cfg, version.Lang, version.Path)
		languageVersions = append(languageVersions, legalDocumentTemplate.LanguageVersion{
			Lang:    version.Lang,
			Path:    version.Path,
			Current: version.Path == doc.Path,
		})
	}

	h.render(w, r.WithContext(ctx), meta, legalDocumentTemplate.ShowPageData{
		Heading:          i18n.T(ctx, page.headingKey),
		HTML:             doc.HTML,
		LanguageVersions: languageVersions,
	})
}

// showVersionは文書の改定前の版をアーカイブページとして表示する
func (h *Handler) showVersion(w http.ResponseWriter, r *http.Request, version legaldoc.Version, doc legaldoc.Document) {
	ctx := i18n.SetLocale(r.Context(), doc.Lang)

	page, ok := documentPages[doc.Key]
	if !ok {
		slog.ErrorContext(ctx, "法的文書のページの設定が登録されていない", "document_key", doc.Key)
		httperror.InternalServerError(w, r.WithContext(ctx))
		return
	}

	effectiveOn := formatDate(ctx, version.EffectiveOn)

	// 検索結果に古い文面が出ないようnoindexにする (canonicalも出力しない)。
	// 現行の版と言語版を結び付けるものではないため、hreflangと言語切替リンクも出力しない
	meta := viewmodel.DefaultPageMeta(ctx, h.cfg, version.Path)
	meta.SetTitle(ctx, "legal_document_show_archive_title", map[string]any{
		"Title":       i18n.T(ctx, page.titleKey),
		"EffectiveOn": effectiveOn,
	})
	meta.Locale = doc.Lang
	meta.NoIndex = true

	h.render(w, r.WithContext(ctx), meta, legalDocumentTemplate.ShowPageData{
		Heading: i18n.T(ctx, "legal_document_show_archive_heading", map[string]any{
			"Heading":     i18n.T(ctx, page.headingKey),
			"EffectiveOn": effectiveOn,
		}),
		HTML: version.HTML,
		ArchiveNotice: &legalDocumentTemplate.ArchiveNotice{
			EffectiveOn: effectiveOn,
			EndedOn:     formatDate(ctx, version.EndedOn),
			CurrentPath: doc.Path,
		},
	})
}

// formatDateは文書の日付をページの言語の書式で表す
func formatDate(ctx context.Context, t time.Time) string {
	return i18n.T(ctx, "legal_document_show_date", map[string]any{
		"Year":      t.Year(),
		"Month":     int(t.Month()),
		"MonthName": t.Month().String(),
		"Day":       t.Day(),
	})
}

// renderは既定レイアウトで法的文書のページを描画する
func (h *Handler) render(w http.ResponseWriter, r *http.Request, meta viewmodel.PageMeta, data legalDocumentTemplate.ShowPageData) {
	ctx := r.Context()

	component := layouts.Default(
		ctx,
		meta,
		viewmodel.NewUserForSidebar(middleware.GetUserFromContext(ctx), h.imageHelper),
		viewmodel.NewSeasons(h.cfg),
		h.cfg.GetAssetVersion(),
		legalDocumentTemplate.Show(data),
	)
	var body bytes.Buffer
	if err := component.Render(ctx, &body); err != nil {
		slog.ErrorContext(ctx, "テンプレートのレンダリングエラー", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write(body.Bytes()); err != nil {
		slog.ErrorContext(ctx, "法的文書のページのレスポンスの書き込みに失敗", "error", err)
	}
}
