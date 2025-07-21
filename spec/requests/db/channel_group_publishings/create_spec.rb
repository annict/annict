# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/channel_groups/:id/publishing", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    channel_group = ChannelGroup.first.tap { |cg| cg.unpublish }

    post "/db/channel_groups/#{channel_group.id}/publishing"
    channel_group.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(channel_group.published?).to eq(false)
  end

  it "エディター権限を持たないユーザーでログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    channel_group = ChannelGroup.first.tap { |cg| cg.unpublish }
    login_as(user, scope: :user)

    post "/db/channel_groups/#{channel_group.id}/publishing"
    channel_group.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(channel_group.published?).to eq(false)
  end

  it "エディター権限を持つユーザーでログインしているとき、アクセスできないこと" do
    user = create(:registered_user, :with_editor_role)
    channel_group = ChannelGroup.first.tap { |cg| cg.unpublish }
    login_as(user, scope: :user)

    post "/db/channel_groups/#{channel_group.id}/publishing"
    channel_group.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(channel_group.published?).to eq(false)
  end

  it "管理者権限を持つユーザーでログインしているとき、チャンネルグループを公開できること" do
    user = create(:registered_user, :with_admin_role)
    channel_group = ChannelGroup.first.tap { |cg| cg.unpublish }
    login_as(user, scope: :user)

    expect(channel_group.published?).to eq(false)

    post "/db/channel_groups/#{channel_group.id}/publishing"
    channel_group.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("公開しました")
    expect(channel_group.published?).to eq(true)
  end
end
