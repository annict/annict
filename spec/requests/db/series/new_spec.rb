# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/series/new", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    get "/db/series/new"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "エディター権限を持たないユーザーがログインしているとき、アクセスが拒否されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/db/series/new"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "エディター権限を持つユーザーがログインしているとき、ページが正常に表示されること" do
    user = create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    get "/db/series/new"

    expect(response.status).to eq(200)
    expect(response.body).to include("シリーズ登録")
  end
end
