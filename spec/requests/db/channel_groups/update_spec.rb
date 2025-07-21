# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /db/channel_groups/:id", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    channel_group = ChannelGroup.first
    old_channel_group = channel_group.attributes
    channel_group_params = {
      name: "ちゃんねるぐるーぷ"
    }

    patch "/db/channel_groups/#{channel_group.id}", params: {channel_group: channel_group_params}
    channel_group.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(channel_group.name).to eq(old_channel_group["name"])
  end

  it "編集者権限を持たないユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    channel_group = ChannelGroup.first
    old_channel_group = channel_group.attributes
    channel_group_params = {
      name: "ちゃんねるぐるーぷ"
    }

    login_as(user, scope: :user)
    patch "/db/channel_groups/#{channel_group.id}", params: {channel_group: channel_group_params}
    channel_group.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(channel_group.name).to eq(old_channel_group["name"])
  end

  it "編集者権限を持つユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user, :with_editor_role)
    channel_group = ChannelGroup.first
    old_channel_group = channel_group.attributes
    channel_group_params = {
      name: "ちゃんねるぐるーぷ"
    }

    login_as(user, scope: :user)
    patch "/db/channel_groups/#{channel_group.id}", params: {channel_group: channel_group_params}
    channel_group.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(channel_group.name).to eq(old_channel_group["name"])
  end

  it "管理者権限を持つユーザーがログインしているとき、チャンネルグループを更新できること" do
    user = create(:registered_user, :with_admin_role)
    channel_group = ChannelGroup.first
    old_channel_group = channel_group.attributes
    channel_group_params = {
      name: "ちゃんねるぐるーぷ"
    }

    login_as(user, scope: :user)

    expect(channel_group.name).to eq(old_channel_group["name"])

    patch "/db/channel_groups/#{channel_group.id}", params: {channel_group: channel_group_params}
    channel_group.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")
    expect(channel_group.name).to eq("ちゃんねるぐるーぷ")
  end

  it "管理者権限を持つユーザーがログインしているとき、並び順を更新できること" do
    user = create(:registered_user, :with_admin_role)
    channel_group = ChannelGroup.first
    channel_group_params = {
      name: channel_group.name,
      sort_number: 999
    }

    login_as(user, scope: :user)
    patch "/db/channel_groups/#{channel_group.id}", params: {channel_group: channel_group_params}
    channel_group.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")
    expect(channel_group.sort_number).to eq(999)
  end

  it "管理者権限を持つユーザーがログインしているとき、空の名前では更新されること" do
    user = create(:registered_user, :with_admin_role)
    channel_group = ChannelGroup.first
    channel_group_params = {
      name: ""
    }

    login_as(user, scope: :user)
    patch "/db/channel_groups/#{channel_group.id}", params: {channel_group: channel_group_params}
    channel_group.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")
    expect(channel_group.name).to eq("")
  end

  it "削除済みのチャンネルグループは更新できないこと" do
    user = create(:registered_user, :with_admin_role)
    channel_group = ChannelGroup.first
    channel_group.destroy!
    channel_group_params = {
      name: "ちゃんねるぐるーぷ"
    }

    login_as(user, scope: :user)

    expect do
      patch "/db/channel_groups/#{channel_group.id}", params: {channel_group: channel_group_params}
    end.to raise_error(ActiveRecord::RecordNotFound)
  end
end
