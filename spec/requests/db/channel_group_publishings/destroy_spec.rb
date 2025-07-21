# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/channel_groups/:id/publishing", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    channel_group = ChannelGroup.first

    delete "/db/channel_groups/#{channel_group.id}/publishing"
    channel_group.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(channel_group.published?).to eq(true)
  end

  it "エディターではないユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    channel_group = ChannelGroup.first
    login_as(user, scope: :user)

    delete "/db/channel_groups/#{channel_group.id}/publishing"
    channel_group.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(channel_group.published?).to eq(true)
  end

  it "エディターロールを持つユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user, :with_editor_role)
    channel_group = ChannelGroup.first
    login_as(user, scope: :user)

    delete "/db/channel_groups/#{channel_group.id}/publishing"
    channel_group.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(channel_group.published?).to eq(true)
  end

  it "管理者ロールを持つユーザーがログインしているとき、チャンネルグループを非公開にできること" do
    user = create(:registered_user, :with_admin_role)
    channel_group = ChannelGroup.first
    login_as(user, scope: :user)

    expect(channel_group.published?).to eq(true)

    delete "/db/channel_groups/#{channel_group.id}/publishing"
    channel_group.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("非公開にしました")
    expect(channel_group.published?).to eq(false)
  end
end
