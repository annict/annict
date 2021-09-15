# frozen_string_literal: true

describe "PATCH /v1/me/records/:id" do
  let(:user) { create(:user, :with_profile, :with_setting) }
  let(:application) { create(:oauth_application, owner: user) }
  let(:access_token) { create(:oauth_access_token, application: application) }
  let(:work) { create(:work, :with_current_season) }
  let(:episode) { create(:episode, work: work) }
  let(:record) { create(:record, :on_episode, work: work, episode: episode, user: user) }
  let(:uniq_comment) { SecureRandom.uuid }

  before do
    data = {
      comment: uniq_comment,
      access_token: access_token.token
    }
    patch api("/v1/me/records/#{record.episode_record.id}", data)
  end

  it "responses 200" do
    expect(response.status).to eq(200)
  end

  it "updates a record" do
    expect(access_token.owner.records.count).to eq(1)
    expect(access_token.owner.records.first.body).to eq(uniq_comment)
    expect(json["comment"]).to eq(uniq_comment)
  end
end
