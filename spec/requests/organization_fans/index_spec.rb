# typed: false
# frozen_string_literal: true

describe "GET /organizations/:organization_id/fans", type: :request do
  let(:organization) { create(:organization) }
  let!(:user) { create(:registered_user) }
  let!(:organization_favorite) { create(:organization_favorite, user: user, organization: organization) }

  it "アクセスできること" do
    get "/organizations/#{organization.id}/fans"

    expect(response.status).to eq(200)
    expect(response.body).to include(organization.name)
    expect(response.body).to include(user.profile.name)
  end
end
