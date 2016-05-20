# frozen_string_literal: true

describe "Api::V1::Me::Records" do
  let(:access_token) { create(:oauth_access_token) }
  let(:work) { create(:work, :with_current_season) }
  let!(:status) do
    create(:status, kind: "watching", work: work, user: access_token.owner)
  end

  describe "GET /v1/me/works" do
    before do
      data = {
        access_token: access_token.token
      }
      get api("/v1/me/works", data)
    end

    it "200が返ること" do
      expect(response.status).to eq(200)
    end

    it "ステータスを更新している作品が取得できること" do
      expect(json["works"][0]["title"]).to eq(work.title)
    end
  end
end
