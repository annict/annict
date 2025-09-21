# typed: false
# frozen_string_literal: true

RSpec.describe "GET /@:username", type: :request do
  it "ログインしているときアクティビティが存在しないとき、EmptyComponentが表示されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/@#{user.username}"

    expect(response.status).to eq(200)
    expect(response.body).to include("アクティビティはありません")
  end

  it "ログインしているときアクティビティが存在するとき、アクティビティが表示されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    status = create(:status, user: user)
    status_activity_group = create(:activity_group, user: user, itemable_type: "Status", single: false)
    create(:activity, user: user, itemable: status, activity_group: status_activity_group)

    record_1 = create(:record, :with_episode_record, user: user)
    episode_record_activity_group = create(:activity_group, user: user, itemable_type: "EpisodeRecord", single: true)
    create(:activity, user: user, itemable: record_1.episode_record, activity_group: episode_record_activity_group)

    record_2 = create(:record, :with_work_record, user: user)
    work_record_activity_group = create(:activity_group, user: user, itemable_type: "WorkRecord", single: true)
    create(:activity, user: user, itemable: record_2.work_record, activity_group: work_record_activity_group)

    get "/@#{user.username}"

    expect(response.status).to eq(200)
    expect(response.body).to include("がステータスを変更しました")
    expect(response.body).to include("が記録しました")
  end

  it "ログインしていないときアクティビティが存在しないとき、EmptyComponentが表示されること" do
    user = create(:registered_user)

    get "/@#{user.username}"

    expect(response.status).to eq(200)
    expect(response.body).to include("アクティビティはありません")
  end

  it "ログインしていないときアクティビティが存在するとき、アクティビティが表示されること" do
    user = create(:registered_user)

    status = create(:status, user: user)
    status_activity_group = create(:activity_group, user: user, itemable_type: "Status", single: false)
    create(:activity, user: user, itemable: status, activity_group: status_activity_group)

    record_1 = create(:record, :with_episode_record, user: user)
    episode_record_activity_group = create(:activity_group, user: user, itemable_type: "EpisodeRecord", single: true)
    create(:activity, user: user, itemable: record_1.episode_record, activity_group: episode_record_activity_group)

    record_2 = create(:record, :with_work_record, user: user)
    work_record_activity_group = create(:activity_group, user: user, itemable_type: "WorkRecord", single: true)
    create(:activity, user: user, itemable: record_2.work_record, activity_group: work_record_activity_group)

    get "/@#{user.username}"

    expect(response.status).to eq(200)
    expect(response.body).to include("がステータスを変更しました")
    expect(response.body).to include("が記録しました")
  end

  it "存在しないユーザー名でアクセスしたとき、404エラーが返されること" do
    get "/@nonexistent_user"

    expect(response.status).to eq(404)
  end

  it "削除されたユーザーにアクセスしたとき、404エラーが返されること" do
    user = create(:registered_user)
    user.update!(deleted_at: Time.current)

    expect {
      get "/@#{user.username}"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "大量のアクティビティが存在するユーザーでも適切にページネーションされること" do
    user = create(:registered_user)

    50.times do
      status = create(:status, user: user)
      activity_group = create(:activity_group, user: user, itemable_type: "Status", single: false)
      create(:activity, user: user, itemable: status, activity_group: activity_group)
    end

    get "/@#{user.username}"
    expect(response.status).to eq(200)

    get "/@#{user.username}?page=2"
    expect(response.status).to eq(200)
  end

  it "無効な文字列をページネーション番号として送信したとき、正常に処理されること" do
    user = create(:registered_user)

    get "/@#{user.username}?page=invalid"

    expect(response.status).to eq(200)
  end

  it "非常に大きなページネーション番号でアクセスしたとき、正常に処理されること" do
    user = create(:registered_user)

    get "/@#{user.username}?page=999999"

    expect(response.status).to eq(200)
  end
end
