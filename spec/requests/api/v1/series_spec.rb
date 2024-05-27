# typed: false
# frozen_string_literal: true

describe "Api::V1::Series" do
  let(:access_token) { create(:oauth_access_token) }
  let!(:series) { create(:series) }

  describe "GET /v1/series" do
    before do
      get api("/v1/series", access_token: access_token.token)
    end

    context "when added no parameters" do
      it "responses 200" do
        expect(response.status).to eq(200)
      end

      it "gets series info" do
        expected_hash = {
          "id" => series.id,
          "name" => series.name,
          "name_ro" => series.name_ro,
          "name_en" => series.name_en
        }
        expect(json["series"][0]).to include(expected_hash)
        expect(json["total_count"]).to eq(1)
        expect(json["next_page"]).to eq(nil)
        expect(json["prev_page"]).to eq(nil)
      end
    end
  end
end
