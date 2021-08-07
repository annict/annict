# frozen_string_literal: true

describe "GET /organizations/:organization_id", type: :request do
  let(:organization) { create(:organization) }

  it "アクセスできること" do
    get "/organizations/#{organization.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include(organization.name)
  end
end
