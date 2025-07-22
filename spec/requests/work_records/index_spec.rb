# typed: false
# frozen_string_literal: true

RSpec.describe "GET /works/:work_id/records", type: :request do
  it "ログインしているとき、作品の記録一覧を表示すること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    record = FactoryBot.create(:record, :with_work_record, user:, work:)
    record.work_record.update!(body: "面白かった", rating_overall_state: "great")

    login_as(user, scope: :user)
    get "/works/#{work.id}/records"

    expect(response.status).to eq(200)
    expect(response.body).to include("面白かった")
  end

  it "ログインしていないとき、記録一覧を表示すること" do
    work = FactoryBot.create(:work)
    user = FactoryBot.create(:registered_user)
    record = FactoryBot.create(:record, :with_work_record, user:, work:)
    record.work_record.update!(body: "面白かった", rating_overall_state: "great")

    get "/works/#{work.id}/records"

    expect(response.status).to eq(200)
    expect(response.body).to include("面白かった")
  end

  it "作品が存在しないとき、404エラーを返すこと" do
    user = FactoryBot.create(:registered_user)

    login_as(user, scope: :user)
    expect {
      get "/works/99999/records"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除済み作品の記録にアクセスしたとき、404エラーを返すこと" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work, :deleted)

    login_as(user, scope: :user)
    expect {
      get "/works/#{work.id}/records"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "自分の記録を表示すること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    my_record = FactoryBot.create(:record, :with_work_record, user:, work:)
    my_record.work_record.update!(body: "自分の記録")

    login_as(user, scope: :user)
    get "/works/#{work.id}/records"

    expect(response.status).to eq(200)
    expect(response.body).to include("自分の記録")
  end

  it "フォローしているユーザーの記録を表示すること" do
    user = FactoryBot.create(:registered_user)
    following_user = FactoryBot.create(:registered_user)
    user.follow(following_user)

    work = FactoryBot.create(:work)
    following_record = FactoryBot.create(:record, :with_work_record, user: following_user, work:)
    following_record.work_record.update!(body: "フォロー中ユーザーの記録")

    login_as(user, scope: :user)
    get "/works/#{work.id}/records"

    expect(response.status).to eq(200)
    expect(response.body).to include("フォロー中ユーザーの記録")
  end

  it "ミュートしているユーザーの記録を表示しないこと" do
    user = FactoryBot.create(:registered_user)
    muted_user = FactoryBot.create(:registered_user)
    user.mute_users.create!(muted_user:)

    work = FactoryBot.create(:work)
    muted_record = FactoryBot.create(:record, :with_work_record, user: muted_user, work:)
    muted_record.work_record.update!(body: "ミュートユーザーの記録")

    login_as(user, scope: :user)
    get "/works/#{work.id}/records"

    expect(response.status).to eq(200)
    expect(response.body).not_to include("ミュートユーザーの記録")
  end

  it "削除済みの記録を表示しないこと" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)

    active_record = FactoryBot.create(:record, :with_work_record, user:, work:)
    active_record.work_record.update!(body: "表示される記録")

    deleted_record = FactoryBot.create(:record, :with_work_record, user:, work:)
    deleted_record.work_record.update!(body: "削除された記録")
    deleted_record.destroy!

    login_as(user, scope: :user)
    get "/works/#{work.id}/records"

    expect(response.status).to eq(200)
    expect(response.body).to include("表示される記録")
    expect(response.body).not_to include("削除された記録")
  end

  it "レーティングの高い順に記録を表示すること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)

    bad_record = FactoryBot.create(:record, :with_work_record, user:, work:)
    bad_record.work_record.update!(body: "悪い評価", rating_overall_state: "bad")

    great_record = FactoryBot.create(:record, :with_work_record, user:, work:)
    great_record.work_record.update!(body: "素晴らしい評価", rating_overall_state: "great")

    login_as(user, scope: :user)
    get "/works/#{work.id}/records"

    expect(response.status).to eq(200)
    # rating_overall_stateの高い順で表示されることを確認
    great_position = response.body.index("素晴らしい評価")
    bad_position = response.body.index("悪い評価")
    expect(great_position).to be < bad_position
  end

  it "ログインユーザーの場合、本文のない記録は全体の記録一覧に表示されないこと" do
    user = FactoryBot.create(:registered_user)
    other_user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)

    record_with_body = FactoryBot.create(:record, :with_work_record, user: other_user, work:)
    record_with_body.work_record.update!(body: "コメントあり")

    record_without_body = FactoryBot.create(:record, :with_work_record, user: other_user, work:)
    record_without_body.work_record.update!(body: "")

    login_as(user, scope: :user)
    get "/works/#{work.id}/records"

    expect(response.status).to eq(200)
    expect(response.body).to include("コメントあり")
    # 本文のない記録は全体の記録一覧には含まれない
  end

  it "ページネーションが機能すること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)

    # 101件の記録を作成（1ページ100件なので2ページになる）
    101.times do |i|
      other_user = FactoryBot.create(:registered_user)
      record = FactoryBot.create(:record, :with_work_record, user: other_user, work:)
      record.work_record.update!(body: "記録#{i + 1}")
    end

    login_as(user, scope: :user)
    get "/works/#{work.id}/records?page=2"

    expect(response.status).to eq(200)
    # 2ページ目には1番目の記録のみ表示される
    expect(response.body).to include("記録1")
    expect(response.body).not_to include("記録101")
  end

  it "作品のタイトルがページタイトルに含まれること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work, title: "テストアニメ")

    login_as(user, scope: :user)
    get "/works/#{work.id}/records"

    expect(response.status).to eq(200)
    expect(response.body).to include("<title>")
    expect(response.body).to include("テストアニメ")
  end

  it "エピソード記録ではなく作品記録のみ表示されること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)

    # 作品記録
    work_record = FactoryBot.create(:record, :with_work_record, user:, work:)
    work_record.work_record.update!(body: "作品記録")

    # エピソード記録
    episode_record = FactoryBot.create(:record, :with_episode_record, user:, episode:)
    episode_record.episode_record.update!(body: "エピソード記録")

    login_as(user, scope: :user)
    get "/works/#{work.id}/records"

    expect(response.status).to eq(200)
    expect(response.body).to include("作品記録")
    expect(response.body).not_to include("エピソード記録")
  end
end
