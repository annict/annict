# typed: false
# frozen_string_literal: true

describe "POST /db/works", type: :request do
  context "user does not sign in" do
    let!(:work_params) do
      {
        title: "作品タイトル"
      }
    end

    it "user can not access this page" do
      post "/db/works", params: {work: work_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(Work.all.size).to eq(0)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:work_params) do
      {
        title: "作品タイトル"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      post "/db/works", params: {work: work_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(Work.all.size).to eq(0)
    end
  end

  context "user who is editor signs in" do
    let!(:number_format) { NumberFormat.first }
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:work_params) do
      {
        title: "作品タイトル",
        title_kana: "タイトル (かな)",
        title_alter: "タイトルの別名",
        title_en: "タイトル (英)",
        title_alter_en: "タイトルの別名 (英)",
        media: "tv",
        official_site_url: "https://example.com",
        official_site_url_en: "https://example.com",
        wikipedia_url: "https://wikipedia.org",
        wikipedia_url_en: "https://wikipedia.org",
        twitter_username: "Twitter",
        twitter_hashtag: "ハッシュタグ",
        sc_tid: 1234,
        mal_anime_id: 5678,
        number_format_id: number_format.id,
        synopsis: "あらすじ",
        synopsis_source: "あらすじのソース",
        synopsis_en: "あらすじ (英)",
        synopsis_source_en: "あらすじのソース (英)",
        season_year: 2020,
        season_name: "winter",
        manual_episodes_count: 1,
        start_episode_raw_number: 1,
        single_episode: false,
        started_on: "2020-01-01",
        ended_on: "2020-03-31"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can create work" do
      expect(Work.all.size).to eq(0)

      post "/db/works", params: {work: work_params}

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("登録しました")

      expect(Work.all.size).to eq(1)
      work = Work.first

      expect(work.title).to eq("作品タイトル")
      expect(work.title_kana).to eq("タイトル (かな)")
      expect(work.title_alter).to eq("タイトルの別名")
      expect(work.title_en).to eq("タイトル (英)")
      expect(work.title_alter_en).to eq("タイトルの別名 (英)")
      expect(work.media).to eq("tv")
      expect(work.official_site_url).to eq("https://example.com")
      expect(work.official_site_url_en).to eq("https://example.com")
      expect(work.wikipedia_url).to eq("https://wikipedia.org")
      expect(work.wikipedia_url_en).to eq("https://wikipedia.org")
      expect(work.twitter_username).to eq("Twitter")
      expect(work.twitter_hashtag).to eq("ハッシュタグ")
      expect(work.sc_tid).to eq(1234)
      expect(work.mal_anime_id).to eq(5678)
      expect(work.number_format_id).to eq(number_format.id)
      expect(work.synopsis).to eq("あらすじ")
      expect(work.synopsis_source).to eq("あらすじのソース")
      expect(work.synopsis_en).to eq("あらすじ (英)")
      expect(work.synopsis_source_en).to eq("あらすじのソース (英)")
      expect(work.season_year).to eq(2020)
      expect(work.season_name).to eq("winter")
      expect(work.manual_episodes_count).to eq(1)
      expect(work.start_episode_raw_number).to eq(1)
      expect(work.single_episode).to eq(false)
      expect(work.started_on.to_s).to eq("2020-01-01")
      expect(work.ended_on.to_s).to eq("2020-03-31")
    end
  end
end
