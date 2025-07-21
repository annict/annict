# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/series", type: :request do
  it "ユーザーがサインインしていないとき、シリーズ一覧が表示されること" do
    series = create(:series)

    get "/db/series"

    expect(response.status).to eq(200)
    expect(response.body).to include(series.name)
  end

  it "ユーザーがサインインしているとき、シリーズ一覧が表示されること" do
    user = create(:registered_user)
    series = create(:series)
    login_as(user, scope: :user)

    get "/db/series"

    expect(response.status).to eq(200)
    expect(response.body).to include(series.name)
  end

  it "削除されたシリーズが表示されないこと" do
    series = create(:series)
    deleted_series = create(:series, deleted_at: Time.current)

    get "/db/series"

    expect(response.status).to eq(200)
    expect(response.body).to include(series.name)
    expect(response.body).not_to include(deleted_series.name)
  end

  it "ページネーションが機能すること" do
    get "/db/series", params: {page: 2}

    expect(response.status).to eq(200)
  end
end
