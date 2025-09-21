# typed: false
# frozen_string_literal: true

RSpec.describe "GET /works/:work_id/info", type: :request do
  it "ログインしていないとき、作品情報ページが表示されること" do
    work = FactoryBot.create(:work)

    get "/works/#{work.id}/info"

    expect(response).to have_http_status(:ok)
    expect(response.body).to include(work.title)
  end

  it "ログインしているとき、作品情報ページが表示されること" do
    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)

    work = FactoryBot.create(:work)

    get "/works/#{work.id}/info"

    expect(response).to have_http_status(:ok)
    expect(response.body).to include(work.title)
  end

  it "削除された作品のとき、404エラーが表示されること" do
    work = FactoryBot.create(:work)
    work.destroy_in_batches

    expect {
      get "/works/#{work.id}/info"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "存在しない作品IDのとき、404エラーが表示されること" do
    get "/works/999999/info"

    expect(response.status).to eq(404)
  end

  it "VODサービスで配信中の番組が表示されること" do
    work = FactoryBot.create(:work)
    channel = FactoryBot.create(:channel, :with_vod)
    FactoryBot.create(:program, work:, channel:, vod_title_code: "test-vod-code")

    get "/works/#{work.id}/info"

    expect(response).to have_http_status(:ok)
    expect(response.body).to include(channel.name)
  end

  it "VODサービスで配信していない番組は表示されないこと" do
    work = FactoryBot.create(:work)
    channel = FactoryBot.create(:channel, :with_vod)
    # vod_title_codeが空の番組
    FactoryBot.create(:program, work:, channel:, vod_title_code: "")

    get "/works/#{work.id}/info"

    expect(response).to have_http_status(:ok)
    expect(response.body).not_to include(channel.name)
  end

  it "削除されたVODチャンネルの番組は表示されないこと" do
    work = FactoryBot.create(:work)
    channel = FactoryBot.create(:channel, :with_vod)
    FactoryBot.create(:program, work:, channel:, vod_title_code: "test-vod-code")
    channel.destroy_in_batches

    get "/works/#{work.id}/info"

    expect(response).to have_http_status(:ok)
    expect(response.body).not_to include("test-vod-code")
  end

  it "複数のVODサービスで配信中の場合、チャンネルのソート順で表示されること" do
    work = FactoryBot.create(:work)
    channel1 = FactoryBot.create(:channel, :with_vod, sort_number: 200)
    channel2 = FactoryBot.create(:channel, :with_vod, sort_number: 100)
    channel3 = FactoryBot.create(:channel, :with_vod, sort_number: 300)

    FactoryBot.create(:program, work:, channel: channel1, vod_title_code: "test-vod-1")
    FactoryBot.create(:program, work:, channel: channel2, vod_title_code: "test-vod-2")
    FactoryBot.create(:program, work:, channel: channel3, vod_title_code: "test-vod-3")

    get "/works/#{work.id}/info"

    expect(response).to have_http_status(:ok)
    # sort_numberの昇順で表示されることを確認
    body = response.body
    channel2_pos = body.index(channel2.name)
    channel1_pos = body.index(channel1.name)
    channel3_pos = body.index(channel3.name)

    expect(channel2_pos).to be < channel1_pos
    expect(channel1_pos).to be < channel3_pos
  end
end
