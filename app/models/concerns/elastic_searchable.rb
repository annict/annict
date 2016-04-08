# frozen_string_literal: true

module ElasticSearchable
  extend ActiveSupport::Concern

  included do
    include Elasticsearch::Model

    SETTINGS = {
      analysis: {
        analyzer: {
          index_analyzer: {
            type: "custom",
            tokenizer: "japanese_normal",
            char_filter: %w(
              normalize
              whitespaces
            ),
            filter: %w(
              lowercase
              trim
              readingform
              asciifolding
              maxlength
              engram
            )
          },
          search_analyzer: {
            type: "custom",
            tokenizer: "japanese_normal",
            char_filter: %w(
              normalize
              whitespaces
              katakana
              romaji
            ),
            filter: %w(
              lowercase
              trim
              maxlength
              readingform
              asciifolding
            )
          }
        },
        char_filter: {
          normalize: {
            "type": "icu_normalizer",
            "name": "nfkc",
            "mode": "compose"
          },
          katakana: {
            type: "mapping",
            mappings: [
              "ぁ=>ァ", "ぃ=>ィ", "ぅ=>ゥ", "ぇ=>ェ", "ぉ=>ォ",
              "っ=>ッ", "ゃ=>ャ", "ゅ=>ュ", "ょ=>ョ",
              "が=>ガ", "ぎ=>ギ", "ぐ=>グ", "げ=>ゲ", "ご=>ゴ",
              "ざ=>ザ", "じ=>ジ", "ず=>ズ", "ぜ=>ゼ", "ぞ=>ゾ",
              "だ=>ダ", "ぢ=>ヂ", "づ=>ヅ", "で=>デ", "ど=>ド",
              "ば=>バ", "び=>ビ", "ぶ=>ブ", "べ=>ベ", "ぼ=>ボ",
              "ぱ=>パ", "ぴ=>ピ", "ぷ=>プ", "ぺ=>ペ", "ぽ=>ポ",
              "ゔ=>ヴ",
              "あ=>ア", "い=>イ", "う=>ウ", "え=>エ", "お=>オ",
              "か=>カ", "き=>キ", "く=>ク", "け=>ケ", "こ=>コ",
              "さ=>サ", "し=>シ", "す=>ス", "せ=>セ", "そ=>ソ",
              "た=>タ", "ち=>チ", "つ=>ツ", "て=>テ", "と=>ト",
              "な=>ナ", "に=>ニ", "ぬ=>ヌ", "ね=>ネ", "の=>ノ",
              "は=>ハ", "ひ=>ヒ", "ふ=>フ", "へ=>ヘ", "ほ=>ホ",
              "ま=>マ", "み=>ミ", "む=>ム", "め=>メ", "も=>モ",
              "や=>ヤ", "ゆ=>ユ", "よ=>ヨ",
              "ら=>ラ", "り=>リ", "る=>ル", "れ=>レ", "ろ=>ロ",
              "わ=>ワ", "を=>ヲ", "ん=>ン"
            ]
          },
          romaji: {
            type: "mapping",
            mappings: [
              "キャ=>kya", "キュ=>kyu", "キョ=>kyo",
              "シャ=>sha", "シュ=>shu", "ショ=>sho",
              "チャ=>cha", "チュ=>chu", "チョ=>cho",
              "ニャ=>nya", "ニュ=>nyu", "ニョ=>nyo",
              "ヒャ=>hya", "ヒュ=>hyu", "ヒョ=>hyo",
              "ミャ=>mya", "ミュ=>myu", "ミョ=>myo",
              "リャ=>rya", "リュ=>ryu", "リョ=>ryo",
              "ファ=>fa", "フィ=>fi", "フェ=>fe", "フォ=>fo",
              "ギャ=>gya", "ギュ=>gyu", "ギョ=>gyo",
              "ジャ=>ja", "ジュ=>ju", "ジョ=>jo",
              "ヂャ=>ja", "ヂュ=>ju", "ヂョ=>jo",
              "ビャ=>bya", "ビュ=>byu", "ビョ=>byo",
              "ヴァ=>va", "ヴィ=>vi", "ヴ=>v", "ヴェ=>ve", "ヴォ=>vo",
              "ァ=>a", "ィ=>i", "ゥ=>u", "ェ=>e", "ォ=>o",
              "ッ=>t",
              "ャ=>ya", "ュ=>yu", "ョ=>yo",
              "ガ=>ga", "ギ=>gi", "グ=>gu", "ゲ=>ge", "ゴ=>go",
              "ザ=>za", "ジ=>ji", "ズ=>zu", "ゼ=>ze", "ゾ=>zo",
              "ダ=>da", "ヂ=>ji", "ヅ=>zu", "デ=>de", "ド=>do",
              "バ=>ba", "ビ=>bi", "ブ=>bu", "ベ=>be", "ボ=>bo",
              "パ=>pa", "ピ=>pi", "プ=>pu", "ペ=>pe", "ポ=>po",
              "ア=>a", "イ=>i", "ウ=>u", "エ=>e", "オ=>o",
              "カ=>ka", "キ=>ki", "ク=>ku", "ケ=>ke", "コ=>ko",
              "サ=>sa", "シ=>shi", "ス=>su", "セ=>se", "ソ=>so",
              "タ=>ta", "チ=>chi", "ツ=>tsu", "テ=>te", "ト=>to",
              "ナ=>na", "ニ=>ni", "ヌ=>nu", "ネ=>ne", "ノ=>no",
              "ハ=>ha", "ヒ=>hi", "フ=>fu", "ヘ=>he", "ホ=>ho",
              "マ=>ma", "ミ=>mi", "ム=>mu", "メ=>me", "モ=>mo",
              "ヤ=>ya", "ユ=>yu", "ヨ=>yo",
              "ラ=>ra", "リ=>ri", "ル=>ru", "レ=>re", "ロ=>ro",
              "ワ=>wa", "ヲ=>o", "ン=>n"
            ]
          },
          whitespaces: {
            type: "pattern_replace",
            pattern: "\\s{2,}",
            replacement: "\u0020"
          }
        },
        filter: {
          readingform: {
            type: "kuromoji_readingform",
            use_romaji: true
          },
          engram: {
            type: "edgeNGram",
            min_gram: 1,
            max_gram: 36
          },
          maxlength: {
            type: "length",
            max: 36
          }
        },
        tokenizer: {
          japanese_normal: {
            mode: "normal",
            type: "kuromoji_tokenizer"
          }
        }
      }
    }
  end
end
