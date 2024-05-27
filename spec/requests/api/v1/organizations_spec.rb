# typed: false
# frozen_string_literal: true

describe "Api::V1::Organizations" do
  let(:access_token) { create(:oauth_access_token) }
  let!(:organization) { create(:organization) }

  describe "GET /v1/organizations" do
    before do
      get api("/v1/organizations", access_token: access_token.token)
    end

    context "when added no parameters" do
      it "responses 200" do
        expect(response.status).to eq(200)
      end

      it "gets organization info" do
        expected_hash = {
          "id" => organization.id,
          "name" => organization.name,
          "name_kana" => organization.name_kana,
          "name_en" => organization.name_en,
          "url" => organization.url,
          "url_en" => organization.url_en,
          "wikipedia_url" => organization.wikipedia_url,
          "wikipedia_url_en" => organization.wikipedia_url_en,
          "twitter_username" => organization.twitter_username,
          "twitter_username_en" => organization.twitter_username_en,
          "favorite_organizations_count" => organization.favorite_users_count,
          "staffs_count" => organization.staffs_count
        }
        expect(json["organizations"][0]).to include(expected_hash)
        expect(json["total_count"]).to eq(1)
        expect(json["next_page"]).to eq(nil)
        expect(json["prev_page"]).to eq(nil)
      end
    end
  end
end
