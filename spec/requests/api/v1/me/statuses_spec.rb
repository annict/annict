# typed: false
# frozen_string_literal: true

RSpec.describe "POST /v1/me/statuses", type: :request do
  it "正常なデータでステータスを作成できること" do
    access_token = create(:oauth_access_token)
    work = create(:work, :with_current_season)

    data = {
      work_id: work.id,
      kind: "wanna_watch",
      access_token: access_token.token
    }
    post api("/v1/me/statuses", data)

    expect(response.status).to eq(204)
    expect(access_token.owner.statuses.count).to eq(1)
    expect(access_token.owner.statuses.first.kind).to eq("wanna_watch")
  end

  it "認証トークンがない場合、エラーが返されること" do
    work = create(:work, :with_current_season)

    data = {
      work_id: work.id,
      kind: "wanna_watch"
    }
    post api("/v1/me/statuses", data)

    expect(response.status).to eq(401)
  end

  it "存在しない作品IDの場合、エラーが返されること" do
    access_token = create(:oauth_access_token)

    data = {
      work_id: "invalid_id",
      kind: "wanna_watch",
      access_token: access_token.token
    }
    post api("/v1/me/statuses", data)

    expect(response.status).to eq(400)
  end

  it "無効なkindの場合、エラーが返されること" do
    access_token = create(:oauth_access_token)
    work = create(:work, :with_current_season)

    data = {
      work_id: work.id,
      kind: "invalid_kind",
      access_token: access_token.token
    }
    post api("/v1/me/statuses", data)

    expect(response.status).to eq(400)
  end

  it "既存のステータスを更新できること" do
    access_token = create(:oauth_access_token)
    work = create(:work, :with_current_season)

    # 最初のステータス作成
    data = {
      work_id: work.id,
      kind: "wanna_watch",
      access_token: access_token.token
    }
    post api("/v1/me/statuses", data)

    expect(response.status).to eq(204)
    expect(access_token.owner.statuses.count).to eq(1)
    expect(access_token.owner.statuses.first.kind).to eq("wanna_watch")

    # ステータス更新
    data = {
      work_id: work.id,
      kind: "watching",
      access_token: access_token.token
    }
    post api("/v1/me/statuses", data)

    expect(response.status).to eq(204)
    expect(access_token.owner.statuses.count).to eq(2)
    expect(access_token.owner.statuses.last.kind).to eq("watching")
  end
end
