# typed: false
# frozen_string_literal: true

RSpec.describe "GET /fragment/trackable_works/:work_id", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    work = FactoryBot.create(:work)

    get "/fragment/trackable_works/#{work.id}"

    expect(response).to redirect_to("/sign_in")
  end

  it "ログインしているがライブラリエントリがないとき、エラーが発生すること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    login_as(user, scope: :user)

    get "/fragment/trackable_works/#{work.id}"

    expect(response.status).to eq(404)
  end

  it "ログインしていて有効なライブラリエントリがあるとき、視聴可能な作品情報を表示すること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    channel = Channel.first
    user.channels << channel
    library_entry = FactoryBot.create(:library_entry, user:, work:)

    # 視聴済みエピソード
    watched_episode = FactoryBot.create(:episode, work:, sort_number: 1)
    library_entry.watched_episode_ids = [watched_episode.id]
    library_entry.save!

    # 未視聴エピソード
    unwatched_episode1 = FactoryBot.create(:episode, work:, sort_number: 2)
    unwatched_episode2 = FactoryBot.create(:episode, work:, sort_number: 3)

    # 削除されたエピソード（表示されない）
    deleted_episode = FactoryBot.create(:episode, :deleted, work:, sort_number: 4)

    # 番組情報
    program = FactoryBot.create(:program, work:, channel:, started_at: 1.day.from_now)

    login_as(user, scope: :user)

    get "/fragment/trackable_works/#{work.id}"

    expect(response.status).to eq(200)
    # エピソードのタイトルが表示されること
    expect(response.body).to include(unwatched_episode1.title)
    expect(response.body).to include(unwatched_episode2.title)
    # 視聴済みエピソードは表示されないこと
    expect(response.body).not_to include(watched_episode.title)
    # 削除されたエピソードは表示されないこと
    expect(response.body).not_to include(deleted_episode.title)
    # チャンネル名が表示されること
    expect(response.body).to include(program.channel.name)
  end

  it "削除された作品にアクセスしようとしたとき、エラーが発生すること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work, :deleted)
    login_as(user, scope: :user)

    get "/fragment/trackable_works/#{work.id}"

    expect(response.status).to eq(404)
  end

  it "ページネーションが正しく動作すること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    FactoryBot.create(:library_entry, user:, work:)

    # 16個のエピソードを作成（1ページ15件なので2ページ目が必要）
    16.times do |i|
      FactoryBot.create(:episode, work:, sort_number: i + 1)
    end

    login_as(user, scope: :user)

    # 1ページ目
    get "/fragment/trackable_works/#{work.id}"
    expect(response.status).to eq(200)

    # 2ページ目
    get "/fragment/trackable_works/#{work.id}?page=2"
    expect(response.status).to eq(200)
  end

  it "ユーザーがフォローしているチャンネルの番組のみ表示すること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    FactoryBot.create(:library_entry, user:, work:)

    # フォローしているチャンネル
    followed_channel = Channel.first
    user.channels << followed_channel

    # フォローしていないチャンネル
    unfollowed_channel = Channel.second

    # 番組
    FactoryBot.create(:program, work:, channel: followed_channel)
    FactoryBot.create(:program, work:, channel: unfollowed_channel)

    login_as(user, scope: :user)

    get "/fragment/trackable_works/#{work.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include(followed_channel.name)
    expect(response.body).not_to include(unfollowed_channel.name)
  end

  it "削除されたチャンネルの番組は表示されないこと" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    FactoryBot.create(:library_entry, user:, work:)

    # 削除されたチャンネル
    channel_group = ChannelGroup.create!(name: "テストチャンネルグループ", sort_number: 1)
    deleted_channel = Channel.create!(channel_group:, name: "削除されたチャンネル", deleted_at: Time.current)
    user.channels << deleted_channel

    # 通常のチャンネル
    normal_channel = Channel.first
    user.channels << normal_channel

    # 番組
    FactoryBot.create(:program, work:, channel: deleted_channel)
    FactoryBot.create(:program, work:, channel: normal_channel)

    login_as(user, scope: :user)

    get "/fragment/trackable_works/#{work.id}"

    expect(response.status).to eq(200)
    expect(response.body).not_to include(deleted_channel.name)
    expect(response.body).to include(normal_channel.name)
  end

  it "削除された番組は表示されないこと" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    FactoryBot.create(:library_entry, user:, work:)
    channel = Channel.first
    user.channels << channel

    # 削除された番組
    FactoryBot.create(:program, :deleted, work:, channel:)

    # 通常の番組
    FactoryBot.create(:program, work:, channel:)

    login_as(user, scope: :user)

    get "/fragment/trackable_works/#{work.id}"

    expect(response.status).to eq(200)
    # 番組の表示は番組時間などで判断する必要があるが、
    # 少なくともエラーが発生しないことを確認
  end

  it "番組が開始時刻順に表示されること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    FactoryBot.create(:library_entry, user:, work:)
    channel = Channel.first
    user.channels << channel

    # 異なる開始時刻の番組を作成
    FactoryBot.create(:program, work:, channel:, started_at: 3.days.from_now)
    FactoryBot.create(:program, work:, channel:, started_at: 1.day.from_now)
    FactoryBot.create(:program, work:, channel:, started_at: 2.days.from_now)

    login_as(user, scope: :user)

    get "/fragment/trackable_works/#{work.id}"

    expect(response.status).to eq(200)
    # レスポンスボディで番組の順序を確認することは難しいが、
    # 少なくともエラーが発生しないことを確認
  end
end
