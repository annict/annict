# frozen_string_literal: true

describe "GET /works/:id", type: :request do
  context "when user does not sign in" do
    let!(:work) { create(:work) }

    it "responses work info" do
      get "/works/#{work.id}"

      expect(response.status).to eq(200)
      expect(response.body).to include(work.title)
    end
  end

  context "when user signs in" do
    let!(:user) { create(:registered_user) }
    let!(:work) { create(:work) }

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
    let!(:work) { create(:work) }
    let!(:trailer) { create(:trailer, work: work) }

    before do
      get "/works/#{work.id}"
    end

    it "displays trailer title" do
      expect(response.status).to eq(200)
      expect(response.body).to include(trailer.title)
    end
  end

  context "when episodes have been added" do
    let!(:work) { create(:work) }
    let!(:episode) { create(:episode, work: work) }

    before do
      get "/works/#{work.id}"
    end

    it "displays episode title" do
      expect(response.status).to eq(200)
      expect(response.body).to include(episode.title)
    end
  end

  context "when vods have been added" do
    let!(:work) { create(:work) }
    let!(:channel) { Channel.with_vod.first }
    let!(:program) { create(:program, work: work, channel: channel, vod_title_code: "xxx") }
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
    let!(:work) { create(:work) }
    let!(:work_record) { create(:work_record) }
    let!(:record) { create(:record, :on_work, work: work, recordable: work_record) }

    before do
      get "/works/#{work.id}"
    end

    it "displays work record body" do
      expect(response.status).to eq(200)
      expect(response.body).to include(record.body)
    end
  end
end
