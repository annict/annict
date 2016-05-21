# frozen_string_literal: true

describe "Api::V1::Episodes" do
  let(:access_token) { create(:oauth_access_token) }
  let!(:work) { create(:work, :with_current_season, :with_episode) }

  describe "GET /v1/episodes" do
    before do
      get api("/v1/episodes", access_token: access_token.token)
    end

    context "パラメータを渡さないとき" do
      it "200が返ること" do
        expect(response.status).to eq(200)
      end

      it "作品情報が取得できること" do
        expect(json["episodes"][0]["title"]).to eq(work.episodes.first.title)
        expect(json["total_count"]).to eq(1)
      end
    end
  end
end
