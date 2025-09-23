# typed: false
# frozen_string_literal: true

RSpec.describe "GET /organizations/:organization_id/fans", type: :request do
  it "組織のファン一覧が表示されること" do
    organization = create(:organization)
    user = create(:registered_user)
    create(:organization_favorite, user:, organization:)

    get "/organizations/#{organization.id}/fans"

    expect(response.status).to eq(200)
    expect(response.body).to include(organization.name)
    expect(response.body).to include(user.profile.name)
  end

  it "削除された組織の場合、404が返されること" do
    organization = create(:organization, deleted_at: Time.current)

    get "/organizations/#{organization.id}/fans"

    expect(response).to have_http_status(:not_found)
  end

  it "削除されたユーザーのファン情報は表示されないこと" do
    organization = create(:organization)
    active_user = create(:registered_user)
    deleted_user = create(:registered_user, deleted_at: Time.current)
    create(:organization_favorite, user: active_user, organization:)
    create(:organization_favorite, user: deleted_user, organization:)

    get "/organizations/#{organization.id}/fans"

    expect(response.status).to eq(200)
    expect(response.body).to include(active_user.profile.name)
    expect(response.body).not_to include(deleted_user.profile.name)
  end

  it "ファンが視聴作品数の多い順にソートされること" do
    organization = create(:organization)
    user1 = create(:registered_user)
    user2 = create(:registered_user)
    create(:organization_favorite, user: user1, organization:, watched_works_count: 10)
    create(:organization_favorite, user: user2, organization:, watched_works_count: 20)

    get "/organizations/#{organization.id}/fans"

    expect(response.status).to eq(200)
    # user2が先に表示されることを確認
    user1_index = response.body.index(user1.profile.name)
    user2_index = response.body.index(user2.profile.name)
    expect(user2_index).to be < user1_index
  end
end
