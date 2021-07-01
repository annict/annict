# frozen_string_literal: true

describe "GET /works/:id", type: :request do
  context "when user does not sign in" do
    let!(:work) { create(:anime) }

    it "responses work info" do
      get "/works/#{work.id}"

      expect(response.status).to eq(200)
      expect(response.body).to include(work.title)
    end
  end

  context "when user signs in" do
    let!(:user) { create(:registered_user) }
    let!(:work) { create(:anime) }

    before do
      login_as(user, scope: :user)
    end

    it "responses series list" do
      get "/works/#{work.id}"

      expect(response.status).to eq(200)
      expect(response.body).to include(work.title)
    end
  end

  context "when trailers are added" do
    let!(:work) { create(:anime) }
    let!(:trailer) { create(:trailer, anime: work) }

    before do
      get "/works/#{work.id}"
    end

    it "displays trailer title" do
      expect(response.status).to eq(200)
      expect(response.body).to include(trailer.title)
    end
  end

  context "when episodes have been added" do
    let!(:work) { create(:anime) }
    let!(:episode) { create(:episode, anime: work) }

    before do
      get "/works/#{work.id}"
    end

    it "displays episode title" do
      expect(response.status).to eq(200)
      expect(response.body).to include(episode.title)
    end
  end

  context "when vods have been added" do
    let!(:work) { create(:anime) }
    let!(:channel) { Channel.with_vod.first }
    let!(:program) { create(:program, anime: work, channel: channel, vod_title_code: "xxx") }
    let!(:vod_title_url) { "https://example.com/#{program.vod_title_code}" }

    before do
      allow_any_instance_of(Program).to receive(:vod_title_url).and_return(vod_title_url)

      get "/works/#{work.id}"
    end

    it "can access to VOD service" do
      expect(response.status).to eq(200)
      expect(response.body).to include(vod_title_url)
    end
  end

  context "when work records have been added" do
    let!(:work) { create(:anime) }
    let!(:record) { create(:record, anime: work) }
    let!(:work_record) { create(:anime_record, anime: work, record: record) }

    before do
      get "/works/#{work.id}"
    end

    it "displays work record body" do
      expect(response.status).to eq(200)
      expect(response.body).to include(work_record.body)
    end
  end
end
