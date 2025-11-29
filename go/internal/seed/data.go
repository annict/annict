package seed

import (
	"fmt"
	"math/rand"
)

// アニメタイトル生成用のデータソース

// 形容詞リスト（アニメタイトルでよく使われる言葉）
var adjectives = []string{
	"魔法の", "伝説の", "秘密の", "不思議な", "輝く",
	"永遠の", "失われた", "禁断の", "幻想の", "運命の",
	"神秘的な", "無敵の", "最強の", "究極の", "聖なる",
	"暗黒の", "光の", "闇の", "炎の", "氷の",
	"風の", "雷の", "水の", "地の", "天の",
	"勇敢な", "孤高の", "美しき", "優しい", "残酷な",
	"小さな", "大きな", "若き", "古の", "新たな",
	"真実の", "偽りの", "希望の", "絶望の", "夢の",
	"星の", "月の", "太陽の", "銀の", "金の",
	"赤い", "青い", "緑の", "黒い", "白い",
}

// 名詞リスト（アニメのテーマやキャラクターに関連する言葉）
var nouns = []string{
	"冒険", "物語", "戦士", "魔女", "勇者",
	"姫", "騎士", "王国", "帝国", "世界",
	"異世界", "学園", "探偵", "忍者", "侍",
	"ドラゴン", "天使", "悪魔", "精霊", "神",
	"剣", "魔法", "呪文", "秘宝", "遺跡",
	"時空", "次元", "銀河", "宇宙", "星",
	"英雄", "伝説", "神話", "運命", "約束",
	"戦争", "平和", "革命", "反逆", "救済",
	"友情", "恋", "絆", "想い", "記憶",
	"学校", "部活", "青春", "日常", "非日常",
	"アイドル", "歌", "音楽", "ダンス", "ライブ",
	"ロボット", "メカ", "AI", "サイボーグ", "機械",
	"吸血鬼", "狼", "猫", "狐", "犬",
	"料理", "カフェ", "レストラン", "美食", "グルメ",
	"スポーツ", "野球", "サッカー", "バスケ", "テニス",
}

// サフィックスリスト（タイトルの末尾でよく使われる言葉）
var suffixes = []string{
	"物語", "伝説", "クロニクル", "サーガ", "戦記",
	"〜始まりの章〜", "〜運命の扉〜", "〜光と闇〜", "〜新世界〜", "〜無限の夢〜",
	"Re:Zero", "異世界転生", "異世界召喚", "タイムリープ", "ループ",
	"第一章", "第二部", "続編", "完結編", "外伝",
	"〜魂の絆〜", "〜永遠の約束〜", "〜運命の歯車〜", "〜星降る夜に〜", "〜希望の光〜",
}

// シーズン名（春夏秋冬）
type SeasonName string

const (
	SeasonSpring SeasonName = "spring"
	SeasonSummer SeasonName = "summer"
	SeasonAutumn SeasonName = "autumn"
	SeasonWinter SeasonName = "winter"
)

// AllSeasons は全シーズンのリスト
var AllSeasons = []SeasonName{
	SeasonSpring,
	SeasonSummer,
	SeasonAutumn,
	SeasonWinter,
}

// メディアタイプ
type MediaType string

const (
	MediaTV    MediaType = "tv"
	MediaOVA   MediaType = "ova"
	MediaMovie MediaType = "movie"
	MediaWeb   MediaType = "web"
)

// AllMediaTypes は全メディアタイプのリスト
var AllMediaTypes = []MediaType{
	MediaTV,
	MediaOVA,
	MediaMovie,
	MediaWeb,
}

// ユーザー名生成用のデータソース
var usernameAdjectives = []string{
	"happy", "sunny", "cool", "brave", "smart",
	"cute", "kind", "fast", "lucky", "magic",
	"sweet", "gentle", "wild", "bright", "dark",
	"red", "blue", "green", "golden", "silver",
	"super", "mega", "ultra", "hyper", "neo",
	"cyber", "digital", "cosmic", "stellar", "lunar",
	"fire", "ice", "wind", "thunder", "earth",
	"mysterious", "legendary", "epic", "awesome", "fantastic",
}

var usernameNouns = []string{
	"cat", "dog", "fox", "wolf", "bear",
	"dragon", "phoenix", "tiger", "lion", "eagle",
	"star", "moon", "sun", "sky", "ocean",
	"hero", "knight", "wizard", "ninja", "samurai",
	"angel", "devil", "spirit", "dream", "hope",
	"gamer", "otaku", "fan", "lover", "master",
	"player", "hunter", "fighter", "warrior", "champion",
	"artist", "creator", "builder", "maker", "designer",
}

// GenerateAnimeTitle はランダムなアニメタイトルを生成します
func GenerateAnimeTitle(r *rand.Rand) string {
	// パターンをランダムに選択
	pattern := r.Intn(4)

	switch pattern {
	case 0:
		// 形容詞 + 名詞（例: 魔法の冒険）
		adj := adjectives[r.Intn(len(adjectives))]
		noun := nouns[r.Intn(len(nouns))]
		return fmt.Sprintf("%s%s", adj, noun)
	case 1:
		// 形容詞 + 名詞 + サフィックス（例: 魔法の冒険物語）
		adj := adjectives[r.Intn(len(adjectives))]
		noun := nouns[r.Intn(len(nouns))]
		suffix := suffixes[r.Intn(len(suffixes))]
		return fmt.Sprintf("%s%s%s", adj, noun, suffix)
	case 2:
		// 名詞のみ（例: 冒険）
		noun := nouns[r.Intn(len(nouns))]
		return noun
	default:
		// 名詞 + サフィックス（例: 冒険物語）
		noun := nouns[r.Intn(len(nouns))]
		suffix := suffixes[r.Intn(len(suffixes))]
		return fmt.Sprintf("%s%s", noun, suffix)
	}
}

// GenerateSeasonYear は2020〜2025年のランダムな年を生成します
func GenerateSeasonYear(r *rand.Rand) int32 {
	// 年は2020〜2025の範囲内のため、int32への変換は安全
	return int32(2020 + r.Intn(6)) // #nosec G115 // 2020〜2025
}

// GenerateSeasonName はランダムなシーズン名を生成します
func GenerateSeasonName(r *rand.Rand) SeasonName {
	return AllSeasons[r.Intn(len(AllSeasons))]
}

// GenerateMediaType はランダムなメディアタイプを生成します
// weightedはtrueの場合、TVアニメの出現率を高くします（よりリアルな分布）
func GenerateMediaType(r *rand.Rand, weighted bool) MediaType {
	if !weighted {
		return AllMediaTypes[r.Intn(len(AllMediaTypes))]
	}

	// 加重ランダム: TV=70%, OVA=10%, Movie=15%, Web=5%
	n := r.Intn(100)
	switch {
	case n < 70:
		return MediaTV
	case n < 80:
		return MediaOVA
	case n < 95:
		return MediaMovie
	default:
		return MediaWeb
	}
}

// GenerateUsername はランダムなユーザー名を生成します
// numberが0の場合は連番なし、1以上の場合は連番を付与します
func GenerateUsername(r *rand.Rand, number int) string {
	adj := usernameAdjectives[r.Intn(len(usernameAdjectives))]
	noun := usernameNouns[r.Intn(len(usernameNouns))]

	if number <= 0 {
		return fmt.Sprintf("%s_%s", adj, noun)
	}
	return fmt.Sprintf("%s_%s_%d", adj, noun, number)
}

// エピソード記録の感想文生成用のデータソース

// 感想文テンプレート（プレースホルダーを含む）
var episodeBodyTemplates = []string{
	"今回の話はとても面白かったです！{character}の活躍が印象的でした。",
	"{scene}のシーンに感動しました。次回も楽しみです。",
	"展開が予想外で驚きました。{emotion}な気持ちになりました。",
	"{character}の成長が感じられる回でした。これからが楽しみです！",
	"{scene}のシーンが素晴らしかったです。{emotion}な展開でした。",
	"今週も面白かった！{character}のセリフが心に残りました。",
	"{emotion}な回でした。{scene}のシーンは特に良かったです。",
	"予想を超える展開で引き込まれました。{character}がかっこよかった！",
	"{scene}の演出が凄かったです。次回も期待しています。",
	"感情移入してしまいました。{character}の気持ちがよく伝わってきました。",
	"{emotion}な話でしたね。{scene}のシーンは忘れられません。",
	"今回は{character}の見せ場が多くて良かったです！",
	"{scene}の作画が素晴らしかったです。{emotion}な気分になりました。",
	"ストーリー展開が気になります。{character}はどうなってしまうのか...",
	"{emotion}な展開でした。{scene}のシーンでは涙が出ました。",
	"今回の話は神回でした！{character}と{scene}のシーンが最高でした。",
	"{scene}のシーンで鳥肌が立ちました。次回が待ち遠しいです。",
	"{character}の決断に共感しました。{emotion}な気持ちになりました。",
	"今週も安定の面白さでした。{scene}のシーンが印象的でしたね。",
	"{emotion}な回でした。{character}の演技も素晴らしかったです。",
	"展開がスピーディーで飽きませんでした。{character}がかわいかった！",
	"{scene}のシーンは何度見ても飽きません。名シーンでした。",
	"今回は{character}の活躍が光っていました。{emotion}な展開でしたね。",
	"{emotion}な話でした。{scene}のシーンの演出が素晴らしかったです。",
	"物語が大きく動きました。{character}の今後が気になります。",
	"今回も期待を裏切らない面白さでした！{scene}のシーンが最高でした。",
	"{character}の魅力が全開でしたね。{emotion}な気分になりました。",
	"{scene}の音楽がマッチしていて感動しました。素晴らしい回でした。",
	"今週も楽しく視聴できました。{character}の表情が良かったです。",
	"{emotion}な展開で続きが気になります。{scene}のシーンは圧巻でした。",
}

// キャラクター関連の言葉
var characterWords = []string{
	"主人公", "ヒロイン", "敵キャラ", "サブキャラ", "ライバル",
	"仲間", "師匠", "先輩", "後輩", "親友",
	"家族", "兄弟", "姉妹", "幼馴染", "転校生",
}

// シーン関連の言葉
var sceneWords = []string{
	"戦闘", "告白", "別れ", "再会", "決戦",
	"日常", "学園", "修行", "旅立ち", "帰還",
	"対決", "共闘", "救出", "邂逅", "覚醒",
}

// 感情関連の言葉
var emotionWords = []string{
	"感動的", "切ない", "嬉しい", "悲しい", "驚き",
	"ワクワク", "ドキドキ", "ハラハラ", "爽やか", "熱い",
	"優しい", "温かい", "楽しい", "面白い", "感慨深い",
}

// GenerateJapaneseEpisodeRecordBody は日本語のエピソード記録感想文を生成します
func GenerateJapaneseEpisodeRecordBody(r *rand.Rand) string {
	// ランダムなテンプレートを選択
	template := episodeBodyTemplates[r.Intn(len(episodeBodyTemplates))]

	// プレースホルダーを置換
	body := template
	body = replacePlaceholder(body, "{character}", characterWords, r)
	body = replacePlaceholder(body, "{scene}", sceneWords, r)
	body = replacePlaceholder(body, "{emotion}", emotionWords, r)

	return body
}

// replacePlaceholder はテンプレート内のプレースホルダーをランダムな単語で置換します
func replacePlaceholder(template, placeholder string, words []string, r *rand.Rand) string {
	if len(words) == 0 {
		return template
	}
	word := words[r.Intn(len(words))]
	return replaceAll(template, placeholder, word)
}

// replaceAll は文字列内のすべてのoldをnewに置換します
func replaceAll(s, old, new string) string {
	result := ""
	for {
		idx := indexOf(s, old)
		if idx == -1 {
			result += s
			break
		}
		result += s[:idx] + new
		s = s[idx+len(old):]
	}
	return result
}

// indexOf は文字列内で最初にsubstrが出現する位置を返します（見つからない場合は-1）
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
