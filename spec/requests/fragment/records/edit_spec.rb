# typed: false
# frozen_string_literal: true

RSpec.describe "GET /fragment/@:username/records/:record_id/edit", type: :request do
  it "未ログインの場合、ログインページにリダイレクトすること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    record = FactoryBot.create(:record, user:, work:)

    get "/fragment/@#{user.username}/records/#{record.id}/edit"

    expect(response).to redirect_to(new_user_session_path)
  end

  it "ユーザーが存在しない場合、404エラーを返すこと" do
    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)

    get "/fragment/@nonexistentuser/records/123/edit"

    expect(response.status).to eq(404)
  end

  it "記録が存在しない場合、404エラーを返すこと" do
    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)

    get "/fragment/@#{user.username}/records/nonexistent/edit"

    expect(response.status).to eq(404)
  end

  it "削除されたユーザーの記録を編集しようとした場合、404エラーを返すこと" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    record = FactoryBot.create(:record, user:, work:)
    username = user.username
    record_id = record.id

    # ユーザーを削除する前にリロードし、その後に削除
    user.reload
    user.destroy!

    another_user = FactoryBot.create(:registered_user)
    login_as(another_user, scope: :user)

    get "/fragment/@#{username}/records/#{record_id}/edit"

    expect(response.status).to eq(404)
  end

  it "削除された記録を編集しようとした場合、404エラーを返すこと" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    record = FactoryBot.create(:record, user:, work:)
    login_as(user, scope: :user)
    record.destroy!

    get "/fragment/@#{user.username}/records/#{record.id}/edit"

    expect(response.status).to eq(404)
  end

  it "他のユーザーの記録を編集しようとした場合、404エラーを返すこと" do
    owner = FactoryBot.create(:registered_user)
    viewer = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    record = FactoryBot.create(:record, user: owner, work:)

    login_as(viewer, scope: :user)

    get "/fragment/@#{owner.username}/records/#{record.id}/edit"

    expect(response.status).to eq(404)
  end

  it "自分の作品記録を編集できること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    record = FactoryBot.create(:record, user:, work:)
    FactoryBot.create(:work_record, user:, work:, record:)

    login_as(user, scope: :user)

    get "/fragment/@#{user.username}/records/#{record.id}/edit"

    expect(response).to have_http_status(:ok)
  end

  it "自分のエピソード記録を編集できること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)
    record = FactoryBot.create(:record, user:, work:)
    FactoryBot.create(:episode_record, user:, work:, episode:, record:)

    login_as(user, scope: :user)

    get "/fragment/@#{user.username}/records/#{record.id}/edit"

    expect(response).to have_http_status(:ok)
  end

  it "show_optionsパラメータがtrueの場合、正しく処理されること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    record = FactoryBot.create(:record, user:, work:)
    FactoryBot.create(:work_record, user:, work:, record:)

    login_as(user, scope: :user)

    get "/fragment/@#{user.username}/records/#{record.id}/edit", params: {show_options: "true"}

    expect(response).to have_http_status(:ok)
  end

  it "show_boxパラメータがtrueの場合、正しく処理されること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    record = FactoryBot.create(:record, user:, work:)
    FactoryBot.create(:work_record, user:, work:, record:)

    login_as(user, scope: :user)

    get "/fragment/@#{user.username}/records/#{record.id}/edit", params: {show_box: "true"}

    expect(response).to have_http_status(:ok)
  end
end
