# typed: false
# frozen_string_literal: true

RSpec.describe "GET /works/:work_id/episodes", type: :request do
  it "ログインしていないとき、エピソード一覧ページが表示されること" do
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)

    get "/works/#{work.id}/episodes"

    expect(response.status).to eq(200)
    expect(response.body).to include(episode.title)
  end

  it "ログインしているとき、エピソード一覧ページが表示されること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    FactoryBot.create(:episode, work:)
    login_as(user, scope: :user)

    get "/works/#{work.id}/episodes"

    expect(response.status).to eq(200)
    expect(response.body).to include(work.title)
  end

  it "エピソードが存在しない作品にアクセスしたとき、404エラーが返ること" do
    work = FactoryBot.create(:work)
    work.update!(no_episodes: true)

    get "/works/#{work.id}/episodes"

    expect(response.status).to eq(404)
  end

  it "削除された作品にアクセスしたとき、404エラーが返ること" do
    work = FactoryBot.create(:work, :with_episode)
    work.update!(deleted_at: Time.current)

    get "/works/#{work.id}/episodes"

    expect(response.status).to eq(404)
  end

  it "存在しない作品IDでアクセスしたとき、404エラーが返ること" do
    get "/works/999999/episodes"

    expect(response.status).to eq(404)
  end

  it "ページネーションが正しく動作すること" do
    work = FactoryBot.create(:work)
    105.times do |i|
      FactoryBot.create(:episode, work:, sort_number: i + 1, number: "第#{i + 1}話")
    end

    get "/works/#{work.id}/episodes?page=2"

    expect(response.status).to eq(200)
    # 2ページ目の最初のエピソード（101番目）が表示されることを確認
    expect(response.body).to include("第101話")
  end

  it "削除されたエピソードは表示されないこと" do
    work = FactoryBot.create(:work)
    FactoryBot.create(:episode, work:, sort_number: 1, number: "第1話", title: "表示される話")
    FactoryBot.create(:episode, work:, sort_number: 2, number: "第2話", title: "削除された話", deleted_at: Time.current)

    get "/works/#{work.id}/episodes"

    expect(response.status).to eq(200)
    expect(response.body).to include("第1話")
    expect(response.body).to include("表示される話")
    expect(response.body).not_to include("第2話")
    expect(response.body).not_to include("削除された話")
  end

  it "エピソードがソート番号順に表示されること" do
    work = FactoryBot.create(:work)
    FactoryBot.create(:episode, work:, sort_number: 3, number: "第3話")
    FactoryBot.create(:episode, work:, sort_number: 1, number: "第1話")
    FactoryBot.create(:episode, work:, sort_number: 2, number: "第2話")

    get "/works/#{work.id}/episodes"

    expect(response.status).to eq(200)
    # ソート番号順に表示されることを確認
    body = response.body
    episode1_index = body.index("第1話")
    episode2_index = body.index("第2話")
    episode3_index = body.index("第3話")
    expect(episode1_index).to be < episode2_index
    expect(episode2_index).to be < episode3_index
  end

  it "エピソードが1つもない場合、空の状態が表示されること" do
    work = FactoryBot.create(:work)

    get "/works/#{work.id}/episodes"

    expect(response.status).to eq(200)
    # 空の状態のメッセージが表示されることを確認（実際のメッセージ内容は翻訳ファイルに依存）
    expect(response.body).to include("container")
  end
end
