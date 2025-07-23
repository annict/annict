# typed: false
# frozen_string_literal: true

RSpec.describe "GET /fragment/episodes/:episode_id/records", type: :request do
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

  it "レーティングの高い順に記録を表示すること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)

    bad_record = FactoryBot.create(:record, :with_episode_record, user:, work:, episode:)
    bad_record.episode_record.update!(body: "悪い評価", rating_state: "bad")

    great_record = FactoryBot.create(:record, :with_episode_record, user:, work:, episode:)
    great_record.episode_record.update!(body: "素晴らしい評価", rating_state: "great")

    login_as(user, scope: :user)
    get "/fragment/episodes/#{episode.id}/records"

    expect(response.status).to eq(200)
    # rating_stateの高い順で表示されることを確認
    great_position = response.body.index("素晴らしい評価")
    bad_position = response.body.index("悪い評価")
    expect(great_position).to be < bad_position
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

  it "ページネーションが機能すること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)

    # 21件の記録を作成（1ページ20件なので2ページになる）
    21.times do |i|
      other_user = FactoryBot.create(:registered_user)
      record = FactoryBot.create(:record, :with_episode_record, user: other_user, work:, episode:)
      record.episode_record.update!(body: "記録#{i + 1}")
    end

    login_as(user, scope: :user)
    get "/fragment/episodes/#{episode.id}/records?page=2"

    expect(response.status).to eq(200)
    # 2ページ目には21番目の記録のみ表示される
    expect(response.body).to include("記録1")
    expect(response.body).not_to include("記録21")
  end
end
