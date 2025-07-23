# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/channel_groups/:id/edit", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    channel_group = ChannelGroup.first

    get "/db/channel_groups/#{channel_group.id}/edit"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "編集者権限を持たないユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    channel_group = ChannelGroup.first
    login_as(user, scope: :user)

    get "/db/channel_groups/#{channel_group.id}/edit"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "編集者権限を持つユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user, :with_editor_role)
    channel_group = ChannelGroup.first
    login_as(user, scope: :user)

    get "/db/channel_groups/#{channel_group.id}/edit"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "管理者権限を持つユーザーがログインしているとき、チャンネルグループ編集フォームが表示されること" do
    user = create(:registered_user, :with_admin_role)
    channel_group = ChannelGroup.first
    login_as(user, scope: :user)

    get "/db/channel_groups/#{channel_group.id}/edit"

    expect(response.status).to eq(200)
    expect(response.body).to include(channel_group.name)
  end
end
