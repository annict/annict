# typed: false
# frozen_string_literal: true

describe "PATCH /db/works/:id", type: :request do
  context "user does not sign in" do
    let!(:work) { create(:work) }
    let!(:old_work) { work.attributes }
    let!(:work_params) do
      {
        title: "タイトルUpdated"
      }
    end

    it "user can not access this page" do
      patch "/db/works/#{work.id}", params: {work: work_params}
      work.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(work.title).to eq(old_work["title"])
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:work) { create(:work) }
    let!(:old_work) { work.attributes }
    let!(:work_params) do
      {
        title: "タイトルUpdated"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      patch "/db/works/#{work.id}", params: {work: work_params}
      work.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(work.title).to eq(old_work["title"])
    end
  end

  context "user who is editor signs in" do
    let!(:number_format) { NumberFormat.first }
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:work) { create(:work) }
    let!(:old_work) { work.attributes }
    let!(:work_params) do
      {
        title: "タイトルUpdated",
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
        no_episodes: false,
        started_on: "2020-01-01",
        ended_on: "2020-03-31"
      }
    end
    let!(:attr_names) { work_params.keys }

    before do
      login_as(user, scope: :user)
    end

    it "user can update work" do
      attr_names.each do |attr_name|
        expect(work.send(attr_name)).to eq(old_work[attr_name.to_s])
      end

      patch "/db/works/#{work.id}", params: {work: work_params}
      work.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("更新しました")

      expect(work.title).to eq("タイトルUpdated")
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
      expect(work.no_episodes).to eq(false)
      expect(work.started_on.to_s).to eq("2020-01-01")
      expect(work.ended_on.to_s).to eq("2020-03-31")
    end
  end
end
