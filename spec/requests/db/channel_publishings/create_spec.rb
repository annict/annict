# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/channels/:id/publishing", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    channel = Channel.first.tap { |c| c.unpublish }

    post "/db/channels/#{channel.id}/publishing"
    channel.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(channel.published?).to eq(false)
  end

  it "編集者権限を持たない一般ユーザーでログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    channel = Channel.first.tap { |c| c.unpublish }
    login_as(user, scope: :user)

    post "/db/channels/#{channel.id}/publishing"
    channel.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(channel.published?).to eq(false)
  end

  it "編集者権限を持つユーザーでログインしているとき、アクセスできないこと" do
    user = create(:registered_user, :with_editor_role)
    channel = Channel.first.tap { |c| c.unpublish }
    login_as(user, scope: :user)

    post "/db/channels/#{channel.id}/publishing"
    channel.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(channel.published?).to eq(false)
  end

  it "管理者権限を持つユーザーでログインしているとき、チャンネルを公開できること" do
    user = create(:registered_user, :with_admin_role)
    channel = Channel.first.tap { |c| c.unpublish }
    login_as(user, scope: :user)

    expect(channel.published?).to eq(false)

    post "/db/channels/#{channel.id}/publishing"
    channel.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("公開しました")
    expect(channel.published?).to eq(true)
  end
end
