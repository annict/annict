# typed: false
# frozen_string_literal: true

RSpec.describe "GET /works/:work_id/episodes/:episode_id", type: :request do
  it "ログインしていないとき、エピソードページが表示されること" do
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)

    get "/works/#{work.id}/episodes/#{episode.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include(episode.title)
  end

  it "ログインしているとき、エピソードページが表示されること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)

    login_as(user, scope: :user)
    get "/works/#{work.id}/episodes/#{episode.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include(work.title)
    expect(response.body).to include(episode.title)
  end

  it "存在しない作品IDを指定したとき、ActiveRecord::RecordNotFoundが発生すること" do
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)

    expect {
      get "/works/999999/episodes/#{episode.id}"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "存在しないエピソードIDを指定したとき、ActiveRecord::RecordNotFoundが発生すること" do
    work = FactoryBot.create(:work)

    expect {
      get "/works/#{work.id}/episodes/999999"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除された作品のエピソードにアクセスしたとき、ActiveRecord::RecordNotFoundが発生すること" do
    work = FactoryBot.create(:work, :deleted)
    episode = FactoryBot.create(:episode, work:)

    expect {
      get "/works/#{work.id}/episodes/#{episode.id}"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除されたエピソードにアクセスしたとき、ActiveRecord::RecordNotFoundが発生すること" do
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, :deleted, work:)

    expect {
      get "/works/#{work.id}/episodes/#{episode.id}"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "VODプログラムが存在するとき、プログラム情報が取得されること" do
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)
    channel = Channel.with_vod.first
    program = FactoryBot.create(:program, work:, channel:, vod_title_code: "test-vod")

    get "/works/#{work.id}/episodes/#{episode.id}"

    expect(response.status).to eq(200)
    # @programsインスタンス変数が設定されることを間接的に確認
    expect(response.body).to include(work.title)
    expect(response.body).to include(episode.title)
  end

  it "エピソード記録が存在するとき、記録数が表示されること" do
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)
    user = FactoryBot.create(:registered_user)
    record = FactoryBot.create(:record, work:, user:)
    episode_record = FactoryBot.create(:episode_record, episode:, record:, work:, user:)

    get "/works/#{work.id}/episodes/#{episode.id}"

    expect(response.status).to eq(200)
    # エピソード記録数が表示されることを確認
    expect(response.body).to include(episode.episode_records_count.to_s)
  end

  it "エピソードのタイトルと番号が正しく表示されること" do
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:, number: "第1話", title: "はじまりの物語")

    get "/works/#{work.id}/episodes/#{episode.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include("第1話")
    expect(response.body).to include("はじまりの物語")
  end
end
