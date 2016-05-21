# frozen_string_literal: true

describe "Api::V1::Me::Programs" do
  let(:access_token) { create(:oauth_access_token) }
  let(:episode) { create(:episode) }
  let!(:status) do
    create(:status, kind: "watching", work: episode.work, user: access_token.owner)
  end
  let(:channel_work) do
    create(:channel_work, user: access_token.owner, work: episode.work)
  end
  let!(:program) { create(:program, episode: episode, channel: channel_work.channel) }

  describe "GET /v1/me/programs" do
    before do
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
