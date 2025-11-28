# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/works", type: :request do
  it "ユーザーがログインしていないとき、作品一覧が表示されること" do
    work = create(:work)

    get "/db/works"

    expect(response.status).to eq(200)
    expect(response.body).to include(work.title)
  end

  it "ユーザーがログインしているとき、作品一覧が表示されること" do
    user = create(:registered_user)
    work = create(:work)
    login_as(user, scope: :user)

    get "/db/works"

    expect(response.status).to eq(200)
    expect(response.body).to include(work.title)
  end

  it "エピソードなしでフィルタリングしたとき、エピソードがない作品のみ表示されること" do
    work_with_episodes = create(:work, :with_episode)
    work_without_episodes = create(:work)

    get "/db/works", params: {no_episodes: "1"}

    expect(response.status).to eq(200)
    expect(response.body).not_to include(work_with_episodes.title)
    expect(response.body).to include(work_without_episodes.title)
  end

  it "画像なしでフィルタリングしたとき、画像がない作品のみ表示されること" do
    work_with_image = create(:work)
    user = create(:registered_user)
    # WorkImageレコードを直接作成
    # Shrineのimage_dataフィールドには、ファイルメタデータのJSON文字列を設定する必要がある
    image_data = {
      id: "test/dummy.jpg",
      storage: "store",
      metadata: {
        filename: "dummy.jpg",
        size: 1024,
        mime_type: "image/jpeg"
      }
    }.to_json
    WorkImage.create!(
      work: work_with_image,
      user: user,
      copyright: "©テスト",
      image_data: image_data
    )
    work_without_image = create(:work)

    get "/db/works", params: {no_image: "1"}

    expect(response.status).to eq(200)
    expect(response.body).not_to include(work_with_image.title)
    expect(response.body).to include(work_without_image.title)
  end

  it "リリースシーズンなしでフィルタリングしたとき、シーズン情報がない作品のみ表示されること" do
    work_with_season = create(:work, season_year: 2023, season_name: "spring")
    work_without_season = create(:work, season_year: nil, season_name: nil)

    get "/db/works", params: {no_release_season: "1"}

    expect(response.status).to eq(200)
    expect(response.body).not_to include(work_with_season.title)
    expect(response.body).to include(work_without_season.title)
  end

  it "スロットなしでフィルタリングしたとき、スロットがない作品のみ表示されること" do
    work_with_slots = create(:work)
    # Slotレコードを直接作成
    Slot.create!(work: work_with_slots, channel_id: 1, started_at: Time.current)
    work_without_slots = create(:work)

    get "/db/works", params: {no_slots: "1"}

    expect(response.status).to eq(200)
    expect(response.body).not_to include(work_with_slots.title)
    expect(response.body).to include(work_without_slots.title)
  end

  it "シーズンでフィルタリングしたとき、指定したシーズンの作品のみ表示されること" do
    spring_work = create(:work, season_year: 2023, season_name: "spring")
    summer_work = create(:work, season_year: 2023, season_name: "summer")
    autumn_work = create(:work, season_year: 2023, season_name: "autumn")

    get "/db/works", params: {season_slugs: ["2023-spring", "2023-summer"]}

    expect(response.status).to eq(200)
    expect(response.body).to include(spring_work.title)
    expect(response.body).to include(summer_work.title)
    expect(response.body).not_to include(autumn_work.title)
  end

  it "削除された作品は表示されないこと" do
    active_work = create(:work, :not_deleted)
    deleted_work = create(:work, :deleted)

    get "/db/works"

    expect(response.status).to eq(200)
    expect(response.body).to include(active_work.title)
    expect(response.body).not_to include(deleted_work.title)
  end

  it "ページネーションが機能すること" do
    # 101件の作品を作成（1ページ100件なので2ページ目が必要）
    101.times do |i|
      create(:work, title: "作品#{i + 1}")
    end

    # 1ページ目の確認
    get "/db/works"
    expect(response.status).to eq(200)
    expect(response.body).to include("作品101") # 最新のものから表示されるため

    # 2ページ目の確認
    get "/db/works", params: {page: 2}
    expect(response.status).to eq(200)
    expect(response.body).to include("作品1") # 最も古いもの
  end

  it "複数のフィルターを組み合わせて使用できること" do
    # エピソードなし & シーズン情報ありの作品
    work1 = create(:work, season_year: 2023, season_name: "spring")

    # エピソードあり & シーズン情報ありの作品
    work2 = create(:work, :with_episode, season_year: 2023, season_name: "spring")

    # エピソードなし & シーズン情報なしの作品
    work3 = create(:work, season_year: nil, season_name: nil)

    get "/db/works", params: {no_episodes: "1", season_slugs: ["2023-spring"]}

    expect(response.status).to eq(200)
    expect(response.body).to include(work1.title)
    expect(response.body).not_to include(work2.title)
    expect(response.body).not_to include(work3.title)
  end
end
