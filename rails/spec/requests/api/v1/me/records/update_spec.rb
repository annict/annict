# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /v1/me/records/:id", type: :request do
  it "正常なデータで記録を更新できること" do
    user = create(:user, :with_profile, :with_setting)
    application = create(:oauth_application, owner: user)
    access_token = create(:oauth_access_token, application: application)
    work = create(:work, :with_current_season)
    episode = create(:episode, work: work)
    record = create(:episode_record, work: work, episode: episode, user: user)
    uniq_comment = SecureRandom.uuid

    data = {
      comment: uniq_comment,
      access_token: access_token.token
    }
    patch api("/v1/me/records/#{record.id}", data)

    expect(response.status).to eq(200)
    expect(access_token.owner.episode_records.count).to eq(1)
    expect(access_token.owner.episode_records.first.body).to eq(uniq_comment)
    expect(json["comment"]).to eq(uniq_comment)
  end

  it "認証トークンがない場合、エラーが返されること" do
    user = create(:user, :with_profile, :with_setting)
    work = create(:work, :with_current_season)
    episode = create(:episode, work: work)
    record = create(:episode_record, work: work, episode: episode, user: user)

    data = {
      comment: "更新されたコメント"
    }
    patch api("/v1/me/records/#{record.id}", data)

    expect(response.status).to eq(401)
  end

  it "存在しない記録IDの場合、エラーが返されること" do
    user = create(:user, :with_profile, :with_setting)
    application = create(:oauth_application, owner: user)
    access_token = create(:oauth_access_token, application: application)

    data = {
      comment: "更新されたコメント",
      access_token: access_token.token
    }
    patch api("/v1/me/records/invalid_id", data)

    expect(response.status).to eq(400)
  end

  it "他のユーザーの記録を更新しようとした場合、エラーが返されること" do
    user1 = create(:user, :with_profile, :with_setting)
    user2 = create(:user, :with_profile, :with_setting)
    application = create(:oauth_application, owner: user2)
    access_token = create(:oauth_access_token, application: application)
    work = create(:work, :with_current_season)
    episode = create(:episode, work: work)
    record = create(:episode_record, work: work, episode: episode, user: user1)

    data = {
      comment: "更新されたコメント",
      access_token: access_token.token
    }
    patch api("/v1/me/records/#{record.id}", data)

    expect(response.status).to eq(404)
  end

  it "無効なデータの場合、エラーが返されること" do
    user = create(:user, :with_profile, :with_setting)
    application = create(:oauth_application, owner: user)
    access_token = create(:oauth_access_token, application: application)
    work = create(:work, :with_current_season)
    episode = create(:episode, work: work)
    record = create(:episode_record, work: work, episode: episode, user: user)

    data = {
      comment: "a" * (Record::MAX_BODY_LENGTH + 1), # 長すぎるコメント
      access_token: access_token.token
    }
    patch api("/v1/me/records/#{record.id}", data)

    expect(response.status).to eq(400)

    expected = {
      errors: [
        {
          type: "invalid_params",
          message: "感想は1048596文字以内で入力してください"
        }
      ]
    }
    expect(json).to include(expected.deep_stringify_keys)
  end
end
