# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/channel_groups/new", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    get "/db/channel_groups/new"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "エディターではないユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/db/channel_groups/new"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "エディターロールを持つユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    get "/db/channel_groups/new"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "管理者ロールを持つユーザーがログインしているとき、ページが表示されること" do
    user = create(:registered_user, :with_admin_role)
    login_as(user, scope: :user)

    get "/db/channel_groups/new"

    expect(response.status).to eq(200)
    expect(response.body).to include("チャンネルグループ登録")
  end
end
