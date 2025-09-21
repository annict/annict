# typed: false
# frozen_string_literal: true

RSpec.describe "GET /@:username/:status_kind", type: :request do
  it "有効なステータス（watching）にアクセスできること" do
    user = create(:registered_user)
    host! ENV.fetch("ANNICT_HOST")

    get "/@#{user.username}/watching"

    expect(response.status).to eq(200)
    expect(response.body).to include("見てる")
  end

  it "有効なステータス（wanna_watch）にアクセスできること" do
    user = create(:registered_user)
    host! ENV.fetch("ANNICT_HOST")

    get "/@#{user.username}/wanna_watch"

    expect(response.status).to eq(200)
    expect(response.body).to include("見たい")
  end

  it "有効なステータス（watched）にアクセスできること" do
    user = create(:registered_user)
    host! ENV.fetch("ANNICT_HOST")

    get "/@#{user.username}/watched"

    expect(response.status).to eq(200)
    expect(response.body).to include("見た")
  end

  it "有効なステータス（on_hold）にアクセスできること" do
    user = create(:registered_user)
    host! ENV.fetch("ANNICT_HOST")

    get "/@#{user.username}/on_hold"

    expect(response.status).to eq(200)
    expect(response.body).to include("一時中断")
  end

  it "有効なステータス（stop_watching）にアクセスできること" do
    user = create(:registered_user)
    host! ENV.fetch("ANNICT_HOST")

    get "/@#{user.username}/stop_watching"

    expect(response.status).to eq(200)
    expect(response.body).to include("視聴中止")
  end

  it "無効なステータスでアクセスした場合、ルーティングエラーになること" do
    user = create(:registered_user)
    host! ENV.fetch("ANNICT_HOST")

    expect {
      get "/@#{user.username}/invalid_status"
    }.to raise_error(ActionController::RoutingError)
  end

  it "存在しないユーザーでアクセスした場合、レコードが見つからないエラーになること" do
    host! ENV.fetch("ANNICT_HOST")

    get "/@nonexistent_user/watching"

    expect(response.status).to eq(404)
  end
end
