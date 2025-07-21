# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/channels/:id/publishing", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    channel = Channel.first

    delete "/db/channels/#{channel.id}/publishing"
    channel.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(channel.published?).to eq(true)
  end

  it "エディター権限を持たないユーザーがログインしているとき、アクセスを拒否すること" do
    user = create(:registered_user)
    channel = Channel.first
    login_as(user, scope: :user)

    delete "/db/channels/#{channel.id}/publishing"
    channel.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(channel.published?).to eq(true)
  end

  it "エディター権限を持つユーザーがログインしているとき、アクセスを拒否すること" do
    user = create(:registered_user, :with_editor_role)
    channel = Channel.first
    login_as(user, scope: :user)

    delete "/db/channels/#{channel.id}/publishing"
    channel.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(channel.published?).to eq(true)
  end

  it "管理者権限を持つユーザーがログインしているとき、チャンネルを非公開にできること" do
    user = create(:registered_user, :with_admin_role)
    channel = Channel.first
    login_as(user, scope: :user)

    expect(channel.published?).to eq(true)

    delete "/db/channels/#{channel.id}/publishing"
    channel.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("非公開にしました")
    expect(channel.published?).to eq(false)
  end

  it "指定されたチャンネルが存在しないとき、404エラーを返すこと" do
    user = create(:registered_user, :with_admin_role)
    login_as(user, scope: :user)

    expect {
      delete "/db/channels/invalid-id/publishing"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "既に非公開のチャンネルに対してリクエストしたとき、404エラーを返すこと" do
    user = create(:registered_user, :with_admin_role)
    channel = Channel.first
    channel.unpublish
    login_as(user, scope: :user)

    expect {
      delete "/db/channels/#{channel.id}/publishing"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除されたチャンネルに対してリクエストしたとき、404エラーを返すこと" do
    user = create(:registered_user, :with_admin_role)
    channel = Channel.first
    channel_id = channel.id
    channel.destroy_in_batches
    login_as(user, scope: :user)

    expect {
      delete "/db/channels/#{channel_id}/publishing"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
