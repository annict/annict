package viewmodel

// Base URL of the Annict space on Wikino, under which both the help pages and the topics
// gathering them live.
//
// [Ja] Wikino上のAnnictスペースのベースURL。ヘルプページも、それらをまとめるトピックも
// この下にある。
const helpSpaceURL = "https://wikino.app/s/annict"

// helpPageURL builds the URL of an Annict help page on Wikino from its page id, keeping the
// address of the help space in one place.
//
// [Ja] helpPageURL は Wikino 上の Annict のヘルプページの URL を、そのページ ID から生成する。
// ヘルプページの置き場所のアドレスを 1 箇所に閉じ込めるためのもの。
func helpPageURL(pageID string) string {
	return helpSpaceURL + "/pages/" + pageID
}

// helpTopicURL builds the URL of a topic in the Annict space on Wikino from its topic number.
// A topic lists the pages under it, so it is the address to hand a reader who is after the
// whole set rather than one page.
//
// [Ja] helpTopicURLはWikino上のAnnictスペースのトピックのURLを、そのトピック番号から生成
// する。トピックは配下のページを一覧するため、1ページではなくまとまりを求める読者に渡すのは
// こちらのアドレスになる。
func helpTopicURL(topicNumber string) string {
	return helpSpaceURL + "/topics/" + topicNumber
}

// Page ids of the help pages the Annict DB forms link to.
//
// TODO: replace the placeholder ids once the help pages are published on Wikino.
//
// [Ja] Annict DB のフォームがリンクするヘルプページのページ ID。
//
// TODO: ヘルプページを Wikino に公開したら、仮の ID を実際のものに差し替える。
const (
	helpWorkEditingPageID       = "xxx"
	helpEpisodeEditingPageID    = "xxx"
	helpEpisodeBulkCreatePageID = "xxx"
)

// Topic number of the topic holding the developer documentation for the Annict APIs
// (authentication, the GraphQL API, and REST API V1).
//
// [Ja] AnnictのAPI (認証・GraphQL API・REST API V1) の開発者向けドキュメントを収めた
// トピックのトピック番号。
const developerHelpTopicNumber = "5"

// HelpWorkEditingURL returns the URL of the help page holding the work editing guideline,
// which the work registration and edit forms link to from under their heading.
//
// [Ja] HelpWorkEditingURL は作品の編集ガイドラインを載せたヘルプページの URL を返す。作品の
// 登録・編集フォームが見出しの下からリンクする先。
func HelpWorkEditingURL() string {
	return helpPageURL(helpWorkEditingPageID)
}

// HelpEpisodeEditingURL returns the URL of the help page holding the episode editing
// guideline, which the episode edit form links to from under its heading.
//
// [Ja] HelpEpisodeEditingURL はエピソードの編集ガイドラインを載せたヘルプページの URL を返す。
// エピソードの編集フォームが見出しの下からリンクする先。
func HelpEpisodeEditingURL() string {
	return helpPageURL(helpEpisodeEditingPageID)
}

// HelpEpisodeBulkCreateURL returns the URL of the help page stating the line format the bulk
// create form expects, which that form links to from under its heading. The form takes free
// text whose columns are positional, and the page itself states neither the order nor the
// shapes a partly filled line takes, so this is where an editor reads them.
//
// [Ja] HelpEpisodeBulkCreateURL は、一括作成フォームが期待する行の形式を述べるヘルプページの
// URL を返す。同フォームが見出しの下からリンクする先。フォームは列が位置で決まる自由入力を
// 受け取るが、画面自身は列の順序も一部だけ入力した行の形も述べないため、編集者がそれらを読む
// のはこのページになる。
func HelpEpisodeBulkCreateURL() string {
	return helpPageURL(helpEpisodeBulkCreatePageID)
}

// DeveloperHelpTopicURL returns the URL of the topic gathering the developer documentation for
// the Annict APIs, which the sidebar links to as its entry point.
//
// [Ja] DeveloperHelpTopicURLはAnnictのAPIの開発者向けドキュメントをまとめたトピックのURLを
// 返す。サイドバーがその入り口としてリンクする先。
func DeveloperHelpTopicURL() string {
	return helpTopicURL(developerHelpTopicNumber)
}
