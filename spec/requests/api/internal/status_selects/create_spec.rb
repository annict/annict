# typed: false
# frozen_string_literal: true

RSpec.describe "POST /api/internal/works/:work_id/status_select", type: :request do
  it "ログインしていないとき、401を返すこと" do
    work = FactoryBot.create(:work)

    post internal_api_work_status_select_path(work_id: work.id), params: {status_kind: "watching"}

    expect(response).to have_http_status(:unauthorized)
  end

  it "ログインしているとき、ステータスを更新して201を返すこと" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    login_as(user, scope: :user)

    post internal_api_work_status_select_path(work_id: work.id), params: {status_kind: "watching"}

    expect(response).to have_http_status(:created)
    expect(response.parsed_body["flash"]["type"]).to eq("notice")
    expect(response.parsed_body["flash"]["message"]).to eq(I18n.t("messages._common.updated"))
  end

  it "存在しないwork_idのとき、404エラーが返されること" do
    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)

    post internal_api_work_status_select_path(work_id: "non-existent-id"), params: {status_kind: "watching"

    expect(response.status).to eq(404)
  end

  it "無効なstatus_kindが渡されたとき、バリデーションエラーで201を返すこと" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    login_as(user, scope: :user)

    post internal_api_work_status_select_path(work_id: work.id), params: {status_kind: "invalid_status"}

    expect(response).to have_http_status(:created)
    expect(response.parsed_body["flash"]["type"]).to eq("notice")
    expect(response.parsed_body["flash"]["message"]).to eq(I18n.t("messages._common.updated"))
  end
end
