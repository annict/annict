# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /v1/me/records/:id", type: :request do
  it "正常に記録を削除できること" do
    user = create(:user, :with_profile, :with_setting)
    application = create(:oauth_application, owner: user)
    access_token = create(:oauth_access_token, application: application)
    work = create(:work, :with_current_season)
    episode = create(:episode, work: work)
    record = create(:episode_record, work: work, episode: episode, user: user)

    expect(access_token.owner.episode_records.count).to eq(1)

    delete api("/v1/me/records/#{record.id}", access_token: access_token.token)

    expect(response.status).to eq(204)
    expect(access_token.owner.episode_records.count).to eq(0)
  end

  it "認証トークンがない場合、エラーが返されること" do
    user = create(:user, :with_profile, :with_setting)
    work = create(:work, :with_current_season)
    episode = create(:episode, work: work)
    record = create(:episode_record, work: work, episode: episode, user: user)

    delete api("/v1/me/records/#{record.id}")

    expect(response.status).to eq(401)
  end

  it "存在しない記録IDの場合、エラーが返されること" do
    user = create(:user, :with_profile, :with_setting)
    application = create(:oauth_application, owner: user)
    access_token = create(:oauth_access_token, application: application)

    delete api("/v1/me/records/invalid_id", access_token: access_token.token)

    expect(response.status).to eq(400)
  end

  it "他のユーザーの記録を削除しようとした場合、エラーが返されること" do
    user1 = create(:user, :with_profile, :with_setting)
    user2 = create(:user, :with_profile, :with_setting)
    application = create(:oauth_application, owner: user2)
    access_token = create(:oauth_access_token, application: application)
    work = create(:work, :with_current_season)
    episode = create(:episode, work: work)
    record = create(:episode_record, work: work, episode: episode, user: user1)

    delete api("/v1/me/records/#{record.id}", access_token: access_token.token)

    expect(response.status).to eq(404)
  end
end
