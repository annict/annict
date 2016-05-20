# frozen_string_literal: true

describe "Api::V1::Works" do
  let(:access_token) { create(:oauth_access_token) }
  let!(:work) { create(:work, :with_current_season) }

  describe "GET /v1/works" do
    before do
      get api("/v1/works", access_token: access_token.token)
    end

    context "パラメータを渡さないとき" do
      it "200が返ること" do
        expect(response.status).to eq(200)
      end

      it "作品情報が取得できること" do
        expect(json["works"][0]["title"]).to eq(work.title)
        expect(json["total_count"]).to eq(1)
      end
    end
  end
end
