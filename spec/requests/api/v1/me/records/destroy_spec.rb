# frozen_string_literal: true

describe "DELETE /v1/me/records/:id" do
  let(:user) { create(:user, :with_profile, :with_setting) }
  let(:application) { create(:oauth_application, owner: user) }
  let(:access_token) { create(:oauth_access_token, application: application) }
  let(:work) { create(:work, :with_current_season) }
  let(:episode) { create(:episode, work: work) }
  let!(:episode_record) { create(:episode_record) }
  let!(:record) { create(:record, work: work, episode: episode, user: user, recordable: episode_record) }

  it "responses 204" do
    delete api("/v1/me/records/#{episode_record.id}", access_token: access_token.token)
    expect(response.status).to eq(204)
  end

  it "deletes a record" do
    expect(access_token.owner.records.count).to eq(1)

    delete api("/v1/me/records/#{episode_record.id}", access_token: access_token.token)

    expect(access_token.owner.records.count).to eq(0)
  end
end
