# frozen_string_literal: true

describe "POST /v1/me/reviews" do
  describe do
    let(:user) { create(:user, :with_profile, :with_setting) }
    let(:application) { create(:oauth_application, owner: user) }
    let(:access_token) { create(:oauth_access_token, application: application) }
    let(:work) { create(:work, :with_current_season) }

    it "creates work record" do
      expect(Record.count).to eq 0
      expect(AnimeRecord.count).to eq 0
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
      expect(AnimeRecord.count).to eq 1
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

      expect(activity_group.itemable_type).to eq "AnimeRecord"
      expect(activity_group.single).to eq true

      expect(activity.activity_group_id).to eq activity_group.id
      expect(activity.itemable).to eq work_record

      expect(json["id"]).to eq work_record.id
      expect(json["body"]).to eq "#{data[:title]}\n\n#{data[:body]}"
      expect(json["rating_overall_state"]).to eq data[:rating_overall_state]
    end
  end

  context "when input data is invalid" do
    let(:user) { create(:user, :with_profile, :with_setting) }
    let(:application) { create(:oauth_application, owner: user) }
    let(:access_token) { create(:oauth_access_token, application: application) }
    let(:work) { create(:work, :with_current_season) }

    it "returns error" do
      data = {
        work_id: work.id,
        body: "あぁ^～心がぴょんぴょんするんじゃぁ^～" * 52_430, # too long body
        access_token: access_token.token
      }
      post api("/v1/me/reviews", data)

      expect(response.status).to eq(400)

      expected = {
        errors: [
          {
            type: "invalid_params",
            message: "本文は1048596文字以内で入力してください"
          }
        ]
      }
      expect(json).to include(expected.deep_stringify_keys)
    end
  end
end
