# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/channel_groups/:id", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    channel_group = ChannelGroup.first

    expect(ChannelGroup.count).to eq(18)

    delete "/db/channel_groups/#{channel_group.id}"
    channel_group.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(ChannelGroup.count).to eq(18)
  end

  it "一般ユーザーでログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    channel_group = ChannelGroup.first
    login_as(user, scope: :user)

    expect(ChannelGroup.count).to eq(18)

    delete "/db/channel_groups/#{channel_group.id}"
    channel_group.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(ChannelGroup.count).to eq(18)
  end

  it "編集者権限を持つユーザーでログインしているとき、アクセスできないこと" do
    user = create(:registered_user, :with_editor_role)
    channel_group = ChannelGroup.first
    login_as(user, scope: :user)

    expect(ChannelGroup.count).to eq(18)

    delete "/db/channel_groups/#{channel_group.id}"
    channel_group.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(ChannelGroup.count).to eq(18)
  end

  it "管理者権限を持つユーザーでログインしているとき、チャンネルグループを論理削除できること" do
    user = create(:registered_user, :with_admin_role)
    channel_group = ChannelGroup.first
    login_as(user, scope: :user)

    expect(ChannelGroup.count).to eq(18)
    expect(channel_group.deleted?).to eq(false)

    delete "/db/channel_groups/#{channel_group.id}"

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("削除しました")
    expect(ChannelGroup.count).to eq(17)
  end

  it "管理者権限を持つユーザーでログインしているとき、存在しないチャンネルグループのIDを指定したときはエラーになること" do
    user = create(:registered_user, :with_admin_role)
    login_as(user, scope: :user)
    non_existent_id = "non-existent-id"

    expect { delete "/db/channel_groups/#{non_existent_id}" }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "管理者権限を持つユーザーでログインしているとき、すでに論理削除されたチャンネルグループは削除できないこと" do
    user = create(:registered_user, :with_admin_role)
    channel_group = create(:channel_group)
    channel_group.update!(deleted_at: Time.current)
    login_as(user, scope: :user)

    expect(channel_group.deleted?).to eq(true)

    expect { delete "/db/channel_groups/#{channel_group.id}" }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
