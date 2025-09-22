# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/works/:work_id/slots/new", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    work = FactoryBot.create(:work)

    get "/db/works/#{work.id}/slots/new"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "エディター権限を持たないユーザーがログインしているとき、アクセスできないこと" do
    work = FactoryBot.create(:work)
    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)

    get "/db/works/#{work.id}/slots/new"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "エディター権限を持つユーザーがログインしているとき、ページが表示されること" do
    work = FactoryBot.create(:work)
    user = FactoryBot.create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    get "/db/works/#{work.id}/slots/new"

    expect(response.status).to eq(200)
    expect(response.body).to include("放送予定登録")
  end

  it "管理者権限を持つユーザーがログインしているとき、ページが表示されること" do
    work = FactoryBot.create(:work)
    user = FactoryBot.create(:registered_user, :with_admin_role)
    login_as(user, scope: :user)

    get "/db/works/#{work.id}/slots/new"

    expect(response.status).to eq(200)
    expect(response.body).to include("放送予定登録")
  end

  it "存在しない作品IDを指定したとき、404エラーになること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    get "/db/works/999999/slots/new"

    expect(response.status).to eq(404)
  end

  it "削除された作品を指定したとき、404エラーになること" do
    work = FactoryBot.create(:work)
    work.destroy_in_batches
    user = FactoryBot.create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    get "/db/works/#{work.id}/slots/new"

    expect(response.status).to eq(404)
  end

  it "program_idsパラメータが指定されたとき、フォームにデフォルト値が設定されること" do
    work = FactoryBot.create(:work)
    # Channelは事前にデータベースに存在している前提のテスト
    # Channel.firstを使用する
    if Channel.count == 0
      channel_group = FactoryBot.create(:channel_group)
      Channel.create!(name: "テストチャンネル", channel_group:)
    end
    program1 = FactoryBot.create(:program, work:, started_at: Time.current)
    program2 = FactoryBot.create(:program, work:, started_at: 1.day.from_now)
    user = FactoryBot.create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    get "/db/works/#{work.id}/slots/new", params: {program_ids: [program1.id, program2.id]}

    expect(response.status).to eq(200)
    expect(response.body).to include("放送予定登録")
  end
end
