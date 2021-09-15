# frozen_string_literal: true

describe "POST /v1/me/records" do
  describe do
    let(:user) { create(:user, :with_profile, :with_setting) }
    let(:application) { create(:oauth_application, owner: user) }
    let(:access_token) { create(:oauth_access_token, application: application) }
    let(:work) { create(:work, :with_current_season) }
    let(:episode) { create(:episode, work: work) }

    it "creates episode record" do
      expect(EpisodeRecord.count).to eq 0
      expect(Record.count).to eq 0
      expect(ActivityGroup.count).to eq 0
      expect(Activity.count).to eq 0

      data = {
        episode_id: episode.id,
        comment: "あぁ^～心がぴょんぴょんするんじゃぁ^～",
        rating: 4.5,
        rating_state: "great",
        access_token: access_token.token
      }
      post api("/v1/me/records", data)

      expect(response.status).to eq(200)

      expect(EpisodeRecord.count).to eq 1
      expect(Record.count).to eq 1
      expect(ActivityGroup.count).to eq 1
      expect(Activity.count).to eq 1

      record = user.records.first
      episode_record = record.episode_record
      activity_group = user.activity_groups.first
      activity = user.activities.first

      expect(record.body).to eq data[:comment]
      expect(record.locale).to eq "ja"
      expect(record.advanced_rating).to eq data[:rating]
      expect(record.rating).to eq data[:rating_state]
      expect(record.episode_id).to eq episode.id
      expect(record.work_id).to eq work.id

      expect(episode_record).not_to be_nil

      expect(activity_group.itemable_type).to eq "Record"
      expect(activity_group.single).to eq true

      expect(activity.activity_group_id).to eq activity_group.id
      expect(activity.itemable).to eq record

      expect(json["id"]).to eq episode_record.id
      expect(json["comment"]).to eq data[:comment]
      expect(json["rating"]).to eq data[:rating]
      expect(json["rating_state"]).to eq data[:rating_state]
    end
  end

  context "when input data is invalid" do
    let(:user) { create(:user, :with_profile, :with_setting) }
    let(:application) { create(:oauth_application, owner: user) }
    let(:access_token) { create(:oauth_access_token, application: application) }
    let(:work) { create(:work, :with_current_season) }
    let(:episode) { create(:episode, work: work) }

    it "returns error" do
      data = {
        episode_id: episode.id,
        comment: "あぁ^～心がぴょんぴょんするんじゃぁ^～" * 52_430, # too long body
        rating: 4.5,
        rating_state: "great",
        access_token: access_token.token
      }
      post api("/v1/me/records", data)

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
end
