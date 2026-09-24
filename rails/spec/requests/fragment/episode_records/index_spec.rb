# typed: false
# frozen_string_literal: true

RSpec.describe "GET /fragment/episodes/:episode_id/records", type: :request do
  # 記録一覧は「自分の記録」「フォロー中のユーザーの記録」「全体の記録」の3つに分かれて描画され、
  # 各記録は `turbo-frame#record_{id}` で囲まれる。
  # 本文の部分一致では別の記録に一致してしまうため、対象一覧の記録IDで検証する。
  def rendered_record_ids(section_index)
    section = Nokogiri::HTML(response.body).css(".c-record-list")[section_index]

    section.css("turbo-frame[id^='record_']").map { |frame| frame["id"].delete_prefix("record_").to_i }
  end

  def my_record_ids
    rendered_record_ids(0)
  end

  def all_record_ids
    rendered_record_ids(2)
  end

  def create_episode_record(user:, episode:, body:, watched_at:, rating_state: nil)
    record = FactoryBot.create(:record, :with_episode_record, user:, work: episode.work, episode:, watched_at:)
    record.episode_record.update!(body:, rating_state:)

    record
  end

  it "ログインしているとき、エピソードの記録一覧を表示すること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)
    record = FactoryBot.create(:record, :with_episode_record, user:, work:, episode:)
    record.episode_record.update!(body: "面白かった", rating_state: "great")

    login_as(user, scope: :user)
    get "/fragment/episodes/#{episode.id}/records"

    expect(response.status).to eq(200)
    expect(response.body).to include("面白かった")
  end

  it "ログインしていないとき、リダイレクトすること" do
    episode = FactoryBot.create(:episode)

    get "/fragment/episodes/#{episode.id}/records"

    expect(response.status).to eq(302)
    expect(response).to redirect_to(new_user_session_path)
  end

  it "エピソードが存在しないとき、404エラーを返すこと" do
    user = FactoryBot.create(:registered_user)

    login_as(user, scope: :user)
    expect {
      get "/fragment/episodes/99999/records"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除済みエピソードの記録にアクセスしたとき、404エラーを返すこと" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, :deleted, work:)

    login_as(user, scope: :user)
    expect {
      get "/fragment/episodes/#{episode.id}/records"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "自分の記録を表示すること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)
    my_record = FactoryBot.create(:record, :with_episode_record, user:, work:, episode:)
    my_record.episode_record.update!(body: "自分の記録")

    login_as(user, scope: :user)
    get "/fragment/episodes/#{episode.id}/records"

    expect(response.status).to eq(200)
    expect(response.body).to include("自分の記録")
  end

  it "フォローしているユーザーの記録を表示すること" do
    user = FactoryBot.create(:registered_user)
    following_user = FactoryBot.create(:registered_user)
    user.follow(following_user)

    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)
    following_record = FactoryBot.create(:record, :with_episode_record, user: following_user, work:, episode:)
    following_record.episode_record.update!(body: "フォロー中ユーザーの記録")

    login_as(user, scope: :user)
    get "/fragment/episodes/#{episode.id}/records"

    expect(response.status).to eq(200)
    expect(response.body).to include("フォロー中ユーザーの記録")
  end

  it "ミュートしているユーザーの記録を表示しないこと" do
    user = FactoryBot.create(:registered_user)
    muted_user = FactoryBot.create(:registered_user)
    user.mute_users.create!(muted_user:)

    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)
    muted_record = FactoryBot.create(:record, :with_episode_record, user: muted_user, work:, episode:)
    muted_record.episode_record.update!(body: "ミュートユーザーの記録")

    login_as(user, scope: :user)
    get "/fragment/episodes/#{episode.id}/records"

    expect(response.status).to eq(200)
    expect(response.body).not_to include("ミュートユーザーの記録")
  end

  it "削除済みの記録を表示しないこと" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)

    active_record = FactoryBot.create(:record, :with_episode_record, user:, work:, episode:)
    active_record.episode_record.update!(body: "表示される記録")

    deleted_record = FactoryBot.create(:record, :with_episode_record, user:, work:, episode:)
    deleted_record.episode_record.update!(body: "削除された記録")
    deleted_record.destroy!

    login_as(user, scope: :user)
    get "/fragment/episodes/#{episode.id}/records"

    expect(response.status).to eq(200)
    expect(response.body).to include("表示される記録")
    expect(response.body).not_to include("削除された記録")
  end

  it "全体の記録一覧を評価の高い順に表示すること" do
    user = FactoryBot.create(:registered_user)
    episode = FactoryBot.create(:episode, work: FactoryBot.create(:work))

    # 視聴日時を評価と逆順にして、並び順が評価で決まることを確かめる
    bad_record = create_episode_record(
      user: FactoryBot.create(:registered_user), episode:,
      body: "悪い評価", rating_state: "bad", watched_at: Time.parse("2026-01-03 12:00:00 +09:00")
    )
    average_record = create_episode_record(
      user: FactoryBot.create(:registered_user), episode:,
      body: "普通の評価", rating_state: "average", watched_at: Time.parse("2026-01-02 12:00:00 +09:00")
    )
    great_record = create_episode_record(
      user: FactoryBot.create(:registered_user), episode:,
      body: "素晴らしい評価", rating_state: "great", watched_at: Time.parse("2026-01-01 12:00:00 +09:00")
    )

    login_as(user, scope: :user)
    get "/fragment/episodes/#{episode.id}/records"

    expect(response.status).to eq(200)
    expect(all_record_ids).to eq([great_record.id, average_record.id, bad_record.id])
  end

  it "自分の記録一覧を視聴日時の新しい順に表示すること" do
    user = FactoryBot.create(:registered_user)
    episode = FactoryBot.create(:episode, work: FactoryBot.create(:work))

    # 評価を視聴日時と逆順にして、自分の記録では評価が並び順に影響しないことを確かめる
    older_record = create_episode_record(
      user:, episode:,
      body: "先に見た記録", rating_state: "great", watched_at: Time.parse("2026-01-01 12:00:00 +09:00")
    )
    newer_record = create_episode_record(
      user:, episode:,
      body: "あとで見た記録", rating_state: "bad", watched_at: Time.parse("2026-01-02 12:00:00 +09:00")
    )

    login_as(user, scope: :user)
    get "/fragment/episodes/#{episode.id}/records"

    expect(response.status).to eq(200)
    expect(my_record_ids).to eq([newer_record.id, older_record.id])
    expect(all_record_ids).to be_empty
  end

  it "本文のない記録は全体の記録一覧に表示されないこと" do
    user = FactoryBot.create(:registered_user)
    other_user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)

    record_with_body = FactoryBot.create(:record, :with_episode_record, user: other_user, work:, episode:)
    record_with_body.episode_record.update!(body: "コメントあり")

    record_without_body = FactoryBot.create(:record, :with_episode_record, user: other_user, work:, episode:)
    record_without_body.episode_record.update!(body: nil)

    login_as(user, scope: :user)
    get "/fragment/episodes/#{episode.id}/records"

    expect(response.status).to eq(200)
    expect(response.body).to include("コメントあり")
    # 本文のない記録は全体の記録一覧には含まれない
  end

  it "全体の記録一覧を1ページ20件でページ送りすること" do
    user = FactoryBot.create(:registered_user)
    episode = FactoryBot.create(:episode, work: FactoryBot.create(:work))
    base_time = Time.parse("2026-01-01 00:00:00 +09:00")

    # 評価の境界がページの境界と一致しないように、高評価11件・低評価10件の計21件を作る
    great_records = Array.new(11) do |i|
      create_episode_record(
        user: FactoryBot.create(:registered_user), episode:,
        body: "高評価の記録#{i + 1}", rating_state: "great", watched_at: base_time + i.hours
      )
    end
    bad_records = Array.new(10) do |i|
      create_episode_record(
        user: FactoryBot.create(:registered_user), episode:,
        body: "低評価の記録#{i + 1}", rating_state: "bad", watched_at: base_time + i.hours
      )
    end
    expected_ids = (great_records.reverse + bad_records.reverse).map(&:id)

    login_as(user, scope: :user)

    get "/fragment/episodes/#{episode.id}/records"

    expect(response.status).to eq(200)
    expect(all_record_ids).to eq(expected_ids.first(20))

    get "/fragment/episodes/#{episode.id}/records?page=2"

    expect(response.status).to eq(200)
    expect(all_record_ids).to eq(expected_ids.last(1))
  end
end
