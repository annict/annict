# typed: false
# frozen_string_literal: true

RSpec.describe "GET /organizations/:organization_id", type: :request do
  it "団体ページにアクセスできること" do
    organization = FactoryBot.create(:organization)

    get "/organizations/#{organization.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include(organization.name)
  end

  it "団体に関連するスタッフ情報が表示されること" do
    organization = FactoryBot.create(:organization)
    work = FactoryBot.create(:work, season_year: 2023)
    FactoryBot.create(:staff, resource: organization, work:)

    get "/organizations/#{organization.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include(work.title)
  end

  it "団体をお気に入りに登録したユーザー情報が表示されること" do
    organization = FactoryBot.create(:organization)
    user = FactoryBot.create(:user)
    FactoryBot.create(:organization_favorite, organization:, user:)

    get "/organizations/#{organization.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include(user.username)
  end

  it "削除済みの作品に関連するスタッフ情報は表示されないこと" do
    organization = FactoryBot.create(:organization)
    work = FactoryBot.create(:work, deleted_at: Time.current)
    FactoryBot.create(:staff, resource: organization, work:)

    get "/organizations/#{organization.id}"

    expect(response.status).to eq(200)
    expect(response.body).not_to include(work.title)
  end

  it "削除済みのスタッフ情報は表示されないこと" do
    organization = FactoryBot.create(:organization)
    work = FactoryBot.create(:work)
    FactoryBot.create(:staff, resource: organization, work:, deleted_at: Time.current)

    get "/organizations/#{organization.id}"

    expect(response.status).to eq(200)
    expect(response.body).not_to include(work.title)
  end

  it "削除済みの団体にアクセスしたときは404エラーが返されること" do
    organization = FactoryBot.create(:organization, deleted_at: Time.current)

    get "/organizations/#{organization.id}"

    expect(response).to have_http_status(:not_found)
  end

  it "存在しない団体IDを指定したときは404エラーが返されること" do
    get "/organizations/999999"

    expect(response).to have_http_status(:not_found)
  end
end
