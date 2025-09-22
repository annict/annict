# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/channels/:id/edit", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    channel = Channel.first

    get "/db/channels/#{channel.id}/edit"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "編集者権限を持つユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user, :with_editor_role)
    channel = Channel.first
    login_as(user, scope: :user)

    get "/db/channels/#{channel.id}/edit"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "管理者権限を持つユーザーがログインしているとき、チャンネル編集フォームが表示されること" do
    user = create(:registered_user, :with_admin_role)
    channel = Channel.first
    login_as(user, scope: :user)

    get "/db/channels/#{channel.id}/edit"

    expect(response.status).to eq(200)
    expect(response.body).to include(channel.name)
  end

  it "一般ユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    channel = Channel.first
    login_as(user, scope: :user)

    get "/db/channels/#{channel.id}/edit"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "削除されたチャンネルにアクセスしようとしたとき、404エラーになること" do
    user = create(:registered_user, :with_admin_role)
    channel = Channel.first
    channel.update!(deleted_at: Time.current)
    login_as(user, scope: :user)

    get "/db/channels/#{channel.id}/edit"

    expect(response.status).to eq(404)
  end

  it "存在しないチャンネルIDでアクセスしたとき、404エラーになること" do
    user = create(:registered_user, :with_admin_role)
    login_as(user, scope: :user)

    get "/db/channels/invalid-id/edit"

    expect(response.status).to eq(404)
  end
end
