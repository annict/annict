# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/works/:work_id/episodes", type: :request do
  it "ユーザーがログインしていないとき、エピソード一覧を表示すること" do
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)

    get "/db/works/#{work.id}/episodes"

    expect(response.status).to eq(200)
    expect(response.body).to include(episode.title)
  end

  it "ユーザーがログインしているとき、エピソード一覧を表示すること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)

    login_as(user, scope: :user)
    get "/db/works/#{work.id}/episodes"

    expect(response.status).to eq(200)
    expect(response.body).to include(episode.title)
  end

  it "複数のエピソードがあるとき、sort_numberの降順で表示すること" do
    work = FactoryBot.create(:work)
    episode1 = FactoryBot.create(:episode, work:, sort_number: 100)
    episode2 = FactoryBot.create(:episode, work:, sort_number: 200)
    episode3 = FactoryBot.create(:episode, work:, sort_number: 150)

    get "/db/works/#{work.id}/episodes"

    expect(response.status).to eq(200)
    # sort_numberの降順なので、episode2, episode3, episode1の順番
    body_index_2 = response.body.index(episode2.title)
    body_index_3 = response.body.index(episode3.title)
    body_index_1 = response.body.index(episode1.title)
    expect(body_index_2).to be < body_index_3
    expect(body_index_3).to be < body_index_1
  end

  it "削除済みのエピソードは表示しないこと" do
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)
    deleted_episode = FactoryBot.create(:episode, :deleted, work:)

    get "/db/works/#{work.id}/episodes"

    expect(response.status).to eq(200)
    expect(response.body).to include(episode.title)
    expect(response.body).not_to include(deleted_episode.title)
  end

  it "削除済みの作品の場合、404エラーが返されること" do
    work = FactoryBot.create(:work, :deleted)

    expect {
      get "/db/works/#{work.id}/episodes"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "存在しない作品IDの場合、404エラーが返されること" do
    expect {
      get "/db/works/99999/episodes"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "ページネーションが機能すること" do
    work = FactoryBot.create(:work)
    # 100件以上のエピソードを作成
    105.times do |i|
      FactoryBot.create(:episode, work:, sort_number: i)
    end

    get "/db/works/#{work.id}/episodes"

    expect(response.status).to eq(200)
    # 最初のページには100件まで表示される
    expect(response.body.scan("<tr>").count - 1).to eq(100) # ヘッダー行を除く

    # 2ページ目
    get "/db/works/#{work.id}/episodes?page=2"

    expect(response.status).to eq(200)
    expect(response.body.scan("<tr>").count - 1).to eq(5) # 残り5件
  end
end
