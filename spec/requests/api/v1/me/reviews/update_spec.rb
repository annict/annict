# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /v1/me/reviews/:id", type: :request do
  it "正常なデータでレビューを更新できること" do
    user = create(:user, :with_profile, :with_setting)
    application = create(:oauth_application, owner: user)
    access_token = create(:oauth_access_token, application: application)
    work = create(:work, :with_current_season)
    record = create(:record, work: work, user: user)
    work_record = create(:work_record, record: record, work: work, user: user)
    uniq_title = SecureRandom.uuid
    uniq_body = SecureRandom.uuid

    data = {
      title: uniq_title,
      body: uniq_body,
      access_token: access_token.token
    }
    patch api("/v1/me/reviews/#{work_record.id}", data)

    expect(response.status).to eq(200)
    expected_body = "#{uniq_title}\n\n#{uniq_body}"
    expect(access_token.owner.work_records.count).to eq(1)
    expect(access_token.owner.work_records.first.body).to eq(expected_body)
    expect(json["body"]).to eq(expected_body)
  end

  it "認証トークンがない場合、エラーが返されること" do
    user = create(:user, :with_profile, :with_setting)
    work = create(:work, :with_current_season)
    record = create(:record, work: work, user: user)
    work_record = create(:work_record, record: record, work: work, user: user)

    data = {
      title: "更新されたタイトル",
      body: "更新されたレビュー"
    }
    patch api("/v1/me/reviews/#{work_record.id}", data)

    expect(response.status).to eq(401)
  end

  it "存在しないレビューIDの場合、エラーが返されること" do
    user = create(:user, :with_profile, :with_setting)
    application = create(:oauth_application, owner: user)
    access_token = create(:oauth_access_token, application: application)

    data = {
      title: "更新されたタイトル",
      body: "更新されたレビュー",
      access_token: access_token.token
    }
    patch api("/v1/me/reviews/invalid_id", data)

    expect(response.status).to eq(400)
  end

  it "他のユーザーのレビューを更新しようとした場合、エラーが返されること" do
    user1 = create(:user, :with_profile, :with_setting)
    user2 = create(:user, :with_profile, :with_setting)
    application = create(:oauth_application, owner: user2)
    access_token = create(:oauth_access_token, application: application)
    work = create(:work, :with_current_season)
    record = create(:record, work: work, user: user1)
    work_record = create(:work_record, record: record, work: work, user: user1)

    data = {
      title: "更新されたタイトル",
      body: "更新されたレビュー",
      access_token: access_token.token
    }
    patch api("/v1/me/reviews/#{work_record.id}", data)

    expect(response.status).to eq(404)
  end

  it "無効なデータの場合、エラーが返されること" do
    user = create(:user, :with_profile, :with_setting)
    application = create(:oauth_application, owner: user)
    access_token = create(:oauth_access_token, application: application)
    work = create(:work, :with_current_season)
    record = create(:record, work: work, user: user)
    work_record = create(:work_record, record: record, work: work, user: user)

    data = {
      body: "a" * (Record::MAX_BODY_LENGTH + 1), # 長すぎるコメント
      access_token: access_token.token
    }
    patch api("/v1/me/reviews/#{work_record.id}", data)

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
