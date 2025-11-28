# typed: false
# frozen_string_literal: true

RSpec.describe "POST /v1/me/reviews", type: :request do
  it "正常なデータでレビューを作成できること" do
    user = create(:user, :with_profile, :with_setting)
    application = create(:oauth_application, owner: user)
    access_token = create(:oauth_access_token, application: application)
    work = create(:work, :with_current_season)

    expect(Record.count).to eq 0
    expect(WorkRecord.count).to eq 0
    expect(ActivityGroup.count).to eq 0
    expect(Activity.count).to eq 0

    data = {
      work_id: work.id,
      title: "あぁ^～心がぴょんぴょんするんじゃぁ^～",
      body: "僕は、リゼちゃん！◯（ ´∀｀ ）◯",
      rating_overall_state: "great",
      access_token: access_token.token
    }
    post api("/v1/me/reviews", data)

    expect(response.status).to eq(200)

    expect(Record.count).to eq 1
    expect(WorkRecord.count).to eq 1
    expect(ActivityGroup.count).to eq 1
    expect(Activity.count).to eq 1

    record = user.records.first
    work_record = user.work_records.first
    activity_group = user.activity_groups.first
    activity = user.activities.first

    expect(record.work_id).to eq work.id

    expect(work_record.body).to eq "#{data[:title]}\n\n#{data[:body]}"
    expect(work_record.locale).to eq "ja"
    expect(work_record.rating_overall_state).to eq data[:rating_overall_state]
    expect(work_record.record_id).to eq record.id
    expect(work_record.work_id).to eq work.id

    expect(activity_group.itemable_type).to eq "WorkRecord"
    expect(activity_group.single).to eq true

    expect(activity.activity_group_id).to eq activity_group.id
    expect(activity.itemable).to eq work_record

    expect(json["id"]).to eq work_record.id
    expect(json["body"]).to eq "#{data[:title]}\n\n#{data[:body]}"
    expect(json["rating_overall_state"]).to eq data[:rating_overall_state]
  end

  it "無効なデータの場合、エラーが返されること" do
    user = create(:user, :with_profile, :with_setting)
    application = create(:oauth_application, owner: user)
    access_token = create(:oauth_access_token, application: application)
    work = create(:work, :with_current_season)

    data = {
      work_id: work.id,
      body: "a" * (Record::MAX_BODY_LENGTH + 1), # 長すぎるコメント
      access_token: access_token.token
    }
    post api("/v1/me/reviews", data)

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

  it "認証トークンがない場合、エラーが返されること" do
    work = create(:work, :with_current_season)

    data = {
      work_id: work.id,
      title: "テストタイトル",
      body: "テストレビュー",
      rating_overall_state: "great"
    }
    post api("/v1/me/reviews", data)

    expect(response.status).to eq(401)
  end

  it "存在しない作品IDの場合、エラーが返されること" do
    user = create(:user, :with_profile, :with_setting)
    application = create(:oauth_application, owner: user)
    access_token = create(:oauth_access_token, application: application)

    data = {
      work_id: "invalid_id",
      title: "テストタイトル",
      body: "テストレビュー",
      rating_overall_state: "great",
      access_token: access_token.token
    }
    post api("/v1/me/reviews", data)

    expect(response.status).to eq(400)
  end
end
