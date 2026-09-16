# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/channel_groups/:id", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    channel_group = create(:channel_group)

    expect {
      delete "/db/channel_groups/#{channel_group.id}"
    }.not_to change(ChannelGroup, :count)

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "一般ユーザーでログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    channel_group = create(:channel_group)
    login_as(user, scope: :user)

    expect {
      delete "/db/channel_groups/#{channel_group.id}"
    }.not_to change(ChannelGroup, :count)

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "編集者権限を持つユーザーでログインしているとき、アクセスできないこと" do
    user = create(:registered_user, :with_editor_role)
    channel_group = create(:channel_group)
    login_as(user, scope: :user)

    expect {
      delete "/db/channel_groups/#{channel_group.id}"
    }.not_to change(ChannelGroup, :count)

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "管理者権限を持つユーザーでログインしているとき、チャンネルグループを削除できること" do
    user = create(:registered_user, :with_admin_role)
    channel_group = create(:channel_group)
    login_as(user, scope: :user)

    expect(channel_group.deleted?).to eq(false)

    expect {
      delete "/db/channel_groups/#{channel_group.id}"
    }.to change(ChannelGroup, :count).by(-1)

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("削除しました")
  end

  it "管理者権限を持つユーザーでログインしているとき、存在しないチャンネルグループのIDを指定したときはエラーになること" do
    user = create(:registered_user, :with_admin_role)
    login_as(user, scope: :user)
    non_existent_id = "non-existent-id"

    expect { delete "/db/channel_groups/#{non_existent_id}" }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "管理者権限を持つユーザーでログインしているとき、すでに論理削除されたチャンネルグループは削除できないこと" do
    user = create(:registered_user, :with_admin_role)
    channel_group = create(:channel_group, :deleted)
    login_as(user, scope: :user)

    expect(channel_group.deleted?).to eq(true)

    expect { delete "/db/channel_groups/#{channel_group.id}" }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
