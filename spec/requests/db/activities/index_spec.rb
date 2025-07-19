# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/activities", type: :request do
  it "ユーザーがログインしていないとき、アクティビティリストが表示されること" do
    db_activity = create(:works_create_activity)
    work = db_activity.trackable

    get "/db/activities"

    expect(response.status).to eq(200)
    expect(response.body).to include(work.title)
  end

  it "ユーザーがログインしているとき、アクティビティリストが表示されること" do
    user = create(:registered_user)
    db_activity = create(:works_create_activity)
    work = db_activity.trackable

    login_as(user, scope: :user)

    get "/db/activities"

    expect(response.status).to eq(200)
    expect(response.body).to include(work.title)
  end

  it "ページネーションが正常に動作すること" do
    create(:works_create_activity)

    get "/db/activities", params: {page: 1}

    expect(response.status).to eq(200)
  end

  it "アクティビティが存在しないとき、空のページが表示されること" do
    get "/db/activities"

    expect(response.status).to eq(200)
  end

  it "無効なページパラメータが指定されたとき、正常に処理されること" do
    create(:works_create_activity)

    get "/db/activities", params: {page: "invalid"}

    expect(response.status).to eq(200)
  end
end
