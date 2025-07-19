# typed: false
# frozen_string_literal: true

RSpec.describe "GET /@:username/favorite_organizations", type: :request do
  it "お気に入り団体がいないとき、アクセスできること" do
    user = FactoryBot.create(:registered_user)

    get "/@#{user.username}/favorite_organizations"

    expect(response.status).to eq(200)
    expect(response.body).to include("団体はありません")
  end

  it "お気に入り団体が存在するとき、団体情報が表示されること" do
    user = FactoryBot.create(:registered_user)
    organization = FactoryBot.create(:organization, :published)
    FactoryBot.create(:organization_favorite, user:, organization:)

    get "/@#{user.username}/favorite_organizations"

    expect(response.status).to eq(200)
    expect(response.body).to include(organization.name)
  end

  it "存在しないユーザー名でアクセスしたとき、RecordNotFoundエラーが発生すること" do
    expect {
      get "/@nonexistent_user/favorite_organizations"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
