package viewmodel

// Wikino上のAnnictスペースのベースURL。ヘルプページも、それらをまとめるトピックも
// この下にある。
const helpSpaceURL = "https://wikino.app/s/annict"

// helpPageURLはWikino上のAnnictのヘルプページのURLを、そのページIDから生成する。
// ヘルプページの置き場所のアドレスを1箇所に閉じ込めるためのもの。
func helpPageURL(pageID string) string {
	return helpSpaceURL + "/pages/" + pageID
}

// helpTopicURLはWikino上のAnnictスペースのトピックのURLを、そのトピック番号から生成
// する。トピックは配下のページを一覧するため、1ページではなくまとまりを求める読者に渡すのは
// こちらのアドレスになる。
func helpTopicURL(topicNumber string) string {
	return helpSpaceURL + "/topics/" + topicNumber
}

// Annict DBのフォームがリンクするヘルプページのページID。
//
// TODO: ヘルプページをWikinoに公開したら、仮のIDを実際のものに差し替える。
const (
	helpWorkEditingPageID       = "xxx"
	helpEpisodeEditingPageID    = "xxx"
	helpEpisodeBulkCreatePageID = "xxx"
)

// AnnictのAPI (認証・GraphQL API・REST API V1) の開発者向けドキュメントを収めた
// トピックのトピック番号。
const developerHelpTopicNumber = "5"

// HelpWorkEditingURLは作品の編集ガイドラインを載せたヘルプページのURLを返す。作品の
// 登録・編集フォームが見出しの下からリンクする先。
func HelpWorkEditingURL() string {
	return helpPageURL(helpWorkEditingPageID)
}

// HelpEpisodeEditingURLはエピソードの編集ガイドラインを載せたヘルプページのURLを返す。
// エピソードの編集フォームが見出しの下からリンクする先。
func HelpEpisodeEditingURL() string {
	return helpPageURL(helpEpisodeEditingPageID)
}

// HelpEpisodeBulkCreateURLは、一括作成フォームが期待する行の形式を述べるヘルプページの
// URLを返す。同フォームが見出しの下からリンクする先。フォームは列が位置で決まる自由入力を
// 受け取るが、画面自身は列の順序も一部だけ入力した行の形も述べないため、編集者がそれらを読む
// のはこのページになる。
func HelpEpisodeBulkCreateURL() string {
	return helpPageURL(helpEpisodeBulkCreatePageID)
}

// DeveloperHelpTopicURLはAnnictのAPIの開発者向けドキュメントをまとめたトピックのURLを
// 返す。サイドバーがその入り口としてリンクする先。
func DeveloperHelpTopicURL() string {
	return helpTopicURL(developerHelpTopicNumber)
}
