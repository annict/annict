# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/channels/new", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    get "/db/channels/new"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "編集者権限を持たないユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    login_as(user, scope: :user)
    get "/db/channels/new"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "編集者権限を持つユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)
    get "/db/channels/new"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "管理者権限を持つユーザーがログインしているとき、ページが表示されること" do
    user = create(:registered_user, :with_admin_role)
    login_as(user, scope: :user)
    get "/db/channels/new"

    expect(response.status).to eq(200)
    expect(response.body).to include("チャンネル登録")
  end
end
