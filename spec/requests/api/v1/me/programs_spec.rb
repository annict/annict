# frozen_string_literal: true

describe "Api::V1::Me::Programs" do
  let(:access_token) { create(:oauth_access_token) }
  let(:work) { create(:work, :with_current_season) }
  let(:episode) { create(:episode, work: work) }
  let!(:program) { create(:program, episode: episode) }

  describe "GET /v1/me/programs" do
    before do
      access_token.owner.receive(program.channel)

      data = {
        access_token: access_token.token
      }
      get api("/v1/me/programs", data)
    end

    it "200が返ること" do
      expect(response.status).to eq(200)
    end

    it "見ている作品の放送予定が取得できること" do
      expect(json["programs"][0]["id"]).to eq(program.id)
    end
  end
end
