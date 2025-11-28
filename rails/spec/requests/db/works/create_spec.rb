# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/works", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    work_params = {
      title: "作品タイトル"
    }

    post "/db/works", params: {work: work_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(Work.all.size).to eq(0)
  end

  it "編集者権限がないユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    work_params = {
      title: "作品タイトル"
    }

    login_as(user, scope: :user)
    post "/db/works", params: {work: work_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(Work.all.size).to eq(0)
  end

  it "編集者権限があるユーザーがログインしているとき、作品を作成できること" do
    number_format = NumberFormat.first
    user = create(:registered_user, :with_editor_role)
    work_params = {
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
      no_episodes: false,
      started_on: "2020-01-01",
      ended_on: "2020-03-31"
    }

    login_as(user, scope: :user)

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
    expect(work.no_episodes).to eq(false)
    expect(work.started_on.to_s).to eq("2020-01-01")
    expect(work.ended_on.to_s).to eq("2020-03-31")
  end

  it "編集者権限があるユーザーがログインしているとき、タイトルが空の場合はバリデーションエラーになること" do
    user = create(:registered_user, :with_editor_role)
    work_params = {
      title: "",
      media: "tv"
    }

    login_as(user, scope: :user)
    post "/db/works", params: {work: work_params}

    expect(response.status).to eq(422)
    expect(Work.all.size).to eq(0)
  end

  it "編集者権限があるユーザーがログインしているとき、メディアが空の場合はバリデーションエラーになること" do
    user = create(:registered_user, :with_editor_role)
    work_params = {
      title: "作品タイトル",
      media: ""
    }

    login_as(user, scope: :user)
    post "/db/works", params: {work: work_params}

    expect(response.status).to eq(422)
    expect(Work.all.size).to eq(0)
  end

  it "編集者権限があるユーザーがログインしているとき、URLフォーマットが不正な場合はバリデーションエラーになること" do
    user = create(:registered_user, :with_editor_role)
    work_params = {
      title: "作品タイトル",
      media: "tv",
      official_site_url: "invalid-url",
      wikipedia_url: "not-a-url"
    }

    login_as(user, scope: :user)
    post "/db/works", params: {work: work_params}

    expect(response.status).to eq(422)
    expect(Work.all.size).to eq(0)
  end

  it "編集者権限があるユーザーがログインしているとき、既存のタイトルと重複する場合はバリデーションエラーになること" do
    user = create(:registered_user, :with_editor_role)
    create(:work, title: "既存の作品")

    work_params = {
      title: "既存の作品",
      media: "tv"
    }

    login_as(user, scope: :user)
    post "/db/works", params: {work: work_params}

    expect(response.status).to eq(422)
    expect(Work.all.size).to eq(1)
  end
end
